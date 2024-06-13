package util

import "io"

// ReadUntilEof reads from r until an error or EOF.
// A successful call returns err == nil, not err == EOF.
func ReadUntilEof(r io.Reader) error {
	b := make([]byte, 512)

	for {
		_, err := r.Read(b)

		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}

	}
}
