package packets

import "errors"

const PING_PACKET = 69
const PONG_PACKET = 420

type DataStream struct {
	Data     []byte
	Length   uint
	Finished bool

	readOffset uint
}

func NewDataStream(data []byte, len uint) *DataStream {
	return &DataStream{
		Data:       data,
		Length:     len,
		readOffset: 0,
	}
}

func (s *DataStream) ReadByte() (byte, error) {
	if s.readOffset < s.Length {
		value := s.Data[s.readOffset]
		s.readOffset += 1
		if s.readOffset >= s.Length {
			s.Finished = true
		}
		return value, nil
	}
	s.Finished = true
	return 0, errors.New("not enough data to read byte")
}

func (s *DataStream) ReadBytes(n uint) ([]byte, error) {
	if s.readOffset+n > s.Length {
		s.Finished = true
		return nil, errors.New("not enough data to read bytes")
	}
	result := s.Data[s.readOffset : s.readOffset+n]
	s.readOffset += n
	if s.readOffset >= s.Length {
		s.Finished = true
	}
	return result, nil
}
