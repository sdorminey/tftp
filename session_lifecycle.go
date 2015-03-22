package main

import "fmt"

type ClientIdentity struct {
	Host string
	Port int
}

type SessionLifecycle struct {
	Sessions map[ClientIdentity]PacketHandler
	Fs       *FileSystem
}

func MakeSessionLifecycle(fs *FileSystem) *SessionLifecycle {
	lifecycle := new(SessionLifecycle)
	lifecycle.Sessions = make(map[ClientIdentity]PacketHandler)
	lifecycle.Fs = fs
	return lifecycle
}

func (s *SessionLifecycle) ProcessPacket(addr ClientIdentity, data []byte) (reply []byte) {
	defer func() {
		// We recover any *ErrorPacket type of panic by terminating the session and
		// forwarding the error to the caller.
		if r := recover(); r != nil {
			fmt.Println("Panic detected:", r)
			switch r.(type) {
			case *ErrorPacket:
			default:
				panic(r)
			}

			existingSession := s.Sessions[addr]
			if existingSession != nil {
				s.TerminateSession(addr)
			}

			packet := r.(*ErrorPacket)
			reply = packet.Marshal()
		}
	}()

	packet := UnmarshalPacket(data)
	fmt.Println("Unmarshalled received packet:", packet)
	replyPacket := s.DispatchPacket(addr, packet)
	fmt.Println("Reply packet:", replyPacket)
	marshalled := MarshalPacket(replyPacket)
	fmt.Println("Marshalled reply packet:", marshalled)
	return marshalled
}

func (s *SessionLifecycle) DispatchPacket(addr ClientIdentity, packet Packet) Packet {
	existingSession := s.Sessions[addr]

	switch packet.(type) {
	case *ReadRequestPacket:
		if existingSession != nil {
			panic(ErrorPacket{ERR_ILLEGAL_OPERATION, "RRQ in progress."})
		}
        readSession := new(ReadSession)
        readSession.Fs = s.Fs
        s.Sessions[addr] = readSession
	case *WriteRequestPacket:
		if existingSession != nil {
			panic(ErrorPacket{ERR_ILLEGAL_OPERATION, "WRQ in progress."})
		}
		writeSession := new(WriteSession)
		writeSession.Fs = s.Fs
		s.Sessions[addr] = writeSession
	default:
		if existingSession == nil {
			panic(ErrorPacket{ERR_ILLEGAL_OPERATION, "No session in progress for you."})
		}
	}

	existingSession = s.Sessions[addr]

	// If we've gotten this far we have a valid session, whether new or existing.
	replyPacket := Dispatch(existingSession, packet)

	fmt.Printf("%v -> %v\n", packet, replyPacket)
	return replyPacket
}

func (s *SessionLifecycle) TerminateSession(addr ClientIdentity) {
	s.Sessions[addr] = nil
}
