// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

package main

import (
	"flag"
	"log"
	"os"
)

var Log = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)

// Todo: strip out panics and use error.
func main() {
	listenPort := flag.Int("port", 69, "port to listen on.")
	host := flag.String("host", "127.0.0.1", "host address to listen on.")
	flag.Parse()

	Log.Printf("Listening on host %s, port %d\n", *host, *listenPort)

	fs := MakeFileSystem()
	Listen(*host, *listenPort, fs)
}
