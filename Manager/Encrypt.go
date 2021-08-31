package Manager

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Encrypt struct {
	inFile     string
	outFile    string
	key        string
	bufferSize int64
}

func (e *Encrypt) Encrypt(inFile string, outFile string, key string, bufferSize int64) error {

	f, err := os.Open(inFile)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	k, err := ioutil.ReadFile(key)

	if err != nil {
		log.Fatal(err)
	}

	b, err := aes.NewCipher(k)

	if err != nil {
		log.Panic(err)
	}

	iv := make([]byte, b.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal(err)
	}

	outfile, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	// The buffer size must be multiple of 16 bytes

	log.Printf("Encrypting backup...")

	buf := make([]byte, bufferSize)
	stream := cipher.NewCTR(b, iv)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			// Write into file
			_, err = outfile.Write(buf[:n])
			if err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read %d bytes: %v", n, err)
			break
		}
	}
	_, err = outfile.Write(iv)
	if err != nil {
		return err
	}

	err = os.Remove(inFile)
	if err != nil {
		return err
	}

	return err
}
