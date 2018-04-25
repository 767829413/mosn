package sofarpc

import (
	"context"
	"reflect"
	log2"log"

	"gitlab.alipay-inc.com/afe/mosn/pkg/log"
	"gitlab.alipay-inc.com/afe/mosn/pkg/types"
)

//All of the protocolMaps

var defaultProtocols = &protocols{
	protocolMaps: make(map[byte]Protocol),
}

type protocols struct {
	protocolMaps map[byte]Protocol
}

func DefaultProtocols() types.Protocols {
	return defaultProtocols
}

func NewProtocols(protocolMaps map[byte]Protocol) types.Protocols {
	return &protocols{
		protocolMaps: protocolMaps,
	}
}

// todo: add error as return value
//PROTOCOL LEVEL's Unified EncodeHeaders for BOLTV1、BOLTV2、TR
func (p *protocols) EncodeHeaders(headers interface{}) (string, types.IoBuffer) {
	var protocolCode byte

	switch headers.(type) {
	case ProtoBasicCmd:
		protocolCode = headers.(ProtoBasicCmd).GetProtocol()
	case map[string]string:
		headersMap := headers.(map[string]string)

		if proto, exist := headersMap[SofaPropertyHeader(HeaderProtocolCode)]; exist {
			protoValue := ConvertPropertyValue(proto, reflect.Uint8)
			protocolCode = protoValue.(byte)
		} else {
			//Codec exception
			log.DefaultLogger.Errorf("Invalid encode headers, should contains 'protocol'")

			return "", nil
		}
	default:
		err := "Invalid encode headers"
		log.DefaultLogger.Debugf(err)

		return "", nil
	}

	log.DefaultLogger.Debugf("[EncodeHeaders]protocol code = ", protocolCode)

	if proto, exists := p.protocolMaps[protocolCode]; exists {
		//Return encoded data in map[string]string to stream layer
		return proto.GetEncoder().EncodeHeaders(headers)
	} else {
		log.DefaultLogger.Errorf("Unknown protocol code: [", protocolCode, "] while encode headers.")

		return "", nil
	}
}

func (p *protocols) EncodeData(data types.IoBuffer) types.IoBuffer {
	return data
}

func (p *protocols) EncodeTrailers(trailers map[string]string) types.IoBuffer {
	return nil
}

func (p *protocols) Decode(context context.Context, data types.IoBuffer, filter types.DecodeFilter) {
	// at least 1 byte for protocol code recognize
	for data.Len() > 1 {
		logger := log.ByContext(context)

		protocolCode := data.Bytes()[0]
		maybeProtocolVersion := data.Bytes()[1]

		logger.Debugf("[Decoder]protocol code = %x, maybeProtocolVersion = %x", protocolCode,  maybeProtocolVersion)

		if proto, exists := p.protocolMaps[protocolCode]; exists {

			//Decode the Binary Streams to Command Type
			if _, cmd := proto.GetDecoder().Decode(context, data); cmd != nil {
				proto.GetCommandHandler().HandleCommand(context, cmd, filter)
			} else {
				break
			}
		} else {
			//Codec Exception
			headers := make(map[string]string, 1)
			headers[types.HeaderException] = types.MosnExceptionCodeC
			logger.Errorf("Unknown protocol code: [%x] while decode in ProtocolDecoder.", protocolCode)

			err := "Unknown protocol code while decode in ProtocolDecoder."
			filter.OnDecodeHeader(GenerateExceptionStreamID(err), headers)

			break
		}
	}
}

func (p *protocols) RegisterProtocol(protocolCode byte, protocol Protocol) {
	if _, exists := p.protocolMaps[protocolCode]; exists {
		log.DefaultLogger.Warnf("protocol alreay Exist:", protocolCode)
	} else {
		p.protocolMaps[protocolCode] = protocol
		if log.DefaultLogger != nil {
			log.DefaultLogger.Debugf("register protocol:", protocolCode)
		} else {
			log2.Println("register protocol:", protocolCode)
		}
	}
}

func (p *protocols) UnRegisterProtocol(protocolCode byte) {
	if _, exists := p.protocolMaps[protocolCode]; exists {
		delete(p.protocolMaps, protocolCode)
		if log.DefaultLogger != nil {
			log.DefaultLogger.Debugf("unregister protocol:", protocolCode)
		} else {
			log2.Println("unregister protocol:", protocolCode)
		}
	}
}

func RegisterProtocol(protocolCode byte, protocol Protocol) {
	defaultProtocols.RegisterProtocol(protocolCode, protocol)
}

func UnRegisterProtocol(protocolCode byte) {
	defaultProtocols.UnRegisterProtocol(protocolCode)
}
