package tcp

import (
	"io"
	"log"

	dissector "github.com/go-gost/tls-dissector"
	"github.com/zensey/transparent-proxy/util"
)

type sniffer struct {
	pw *io.PipeWriter
}

func (p *sniffer) close() {
	if p.pw != nil {
		p.pw.Close()
	}
}

func (p *sniffer) capture(src io.Reader, sn *string) io.Reader {
	pr, pw := io.Pipe()
	p.pw = pw
	srcCopy := io.TeeReader(src, pw)

	go func() {
		// must read the stream until the end in any case
		// due to synchronous nature of TeeReader
		defer util.ReadUntilEof(pr)

		rec := Record{}
		_, err := rec.ReadFrom(pr)
		if err != nil {
			log.Println("ReadFrom err:", err)
			return
		}
		if !rec.Valid() {
			// TLS record is not found
			return
		}

		clientHello := dissector.ClientHelloHandshake{}
		_, err = clientHello.ReadFrom(pr)
		if err != nil {
			log.Println("ReadFrom err:", err)
			return
		}

		for _, ext := range clientHello.Extensions {
			if ext.Type() == dissector.ExtServerName {
				snExtension := ext.(*dissector.ServerNameExtension)
				*sn = snExtension.Name
				break
			}
		}
	}()

	return srcCopy
}
