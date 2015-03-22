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
	fmt.Printf(
		"Received %d packet with length %d\n",
		ConvertToUInt16(data[0:2]),
		len(data))

	packet := UnmarshalPacket(data)
	replyPacket := s.DispatchPacket(addr, packet)
	if replyPacket == nil {
		// No response
		return nil
	}

	// Terminate session on ERROR returned.
	_, hasError := replyPacket.(*ErrorPacket)
	if hasError {
		s.TerminateSession(addr)
	}

	marshalled := MarshalPacket(replyPacket)

	fmt.Printf(
		"Sent %d packet with length %d",
		ConvertToUInt16(marshalled[0:2]),
		len(marshalled))

	return marshalled
}

func (s *SessionLifecycle) DispatchPacket(addr ClientIdentity, packet Packet) Packet {
	existingSession := s.Sessions[addr]

	switch packet.(type) {
	case *ReadRequestPacket:
		if existingSession != nil {
			return &ErrorPacket{ERR_ILLEGAL_OPERATION, "RRQ in progress."}
		}
		readSession := MakeReadSession(s.Fs)
		s.Sessions[addr] = readSession
	case *WriteRequestPacket:
		if existingSession != nil {
			return &ErrorPacket{ERR_ILLEGAL_OPERATION, "WRQ in progress."}
		}
		writeSession := MakeWriteSession(s.Fs)
		s.Sessions[addr] = writeSession
	default:
		if existingSession == nil {
			return &ErrorPacket{ERR_ILLEGAL_OPERATION, "Unknown session!"}
		}
	}

	existingSession = s.Sessions[addr]

	// If we've gotten this far we have a valid session, whether new or existing.
	replyPacket := Dispatch(existingSession, packet)

	if existingSession.WantsToDie() {
		s.TerminateSession(addr)
	}

	return replyPacket
}

func (s *SessionLifecycle) TerminateSession(addr ClientIdentity) {
	s.Sessions[addr] = nil
}
