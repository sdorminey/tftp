package main

type ClientIdentity struct {
    Host string
    Port int
}

type SessionLifecycle struct {
    Sessions map[ClientIdentity]PacketHandler
}

func (s *SessionLifecycle) ProcessPacket(addr ClientIdentity, data []byte) (reply []byte) {
    defer func() {
        // We recover any *ErrorPacket type of panic by terminating the session and
        // forwarding the error to the caller.
        if r := recover(); r != nil {
            existingSession := s.Sessions[addr]
            if existingSession != nil {
                switch r.(type) {
                case *ErrorPacket:
                    s.TerminateSession(addr)
                    packet := r.(*ErrorPacket)
                    reply = packet.Marshal()
                default:
                    panic(r)
                }
            }
        }
    }()

    packet := UnmarshalPacket(data)
    replyPacket := s.DispatchPacket(addr, packet)
    return MarshalPacket(replyPacket)
}

func (s *SessionLifecycle) DispatchPacket(addr ClientIdentity, packet Packet) Packet {
    existingSession := s.Sessions[addr]

    switch packet.(type) {
        case *ReadRequestPacket:
            if existingSession != nil {
                panic(ErrorPacket{ERR_ILLEGAL_OPERATION, "RRQ in progress."})
            }
            panic("Not implemented yet")
            //s.Sessions[addr] = new(ReadSession)
        case *WriteRequestPacket:
            if existingSession != nil {
                panic(ErrorPacket{ERR_ILLEGAL_OPERATION, "WRQ in progress."})
            }
            s.Sessions[addr] = new(WriteSession)
        default:
            if existingSession == nil {
                panic(ErrorPacket{ERR_ILLEGAL_OPERATION, "No session in progress for you."})
            }
    }

    existingSession = s.Sessions[addr]

    // If we've gotten this far we have a valid session, whether new or existing.
    return Dispatch(existingSession, packet)
}

func (s *SessionLifecycle) TerminateSession(addr ClientIdentity) {
    s.Sessions[addr] = nil
}
