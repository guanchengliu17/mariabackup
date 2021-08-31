package Manager

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Decrypt struct {
	inFile     string
	outFile    string
	key        string
	bufferSize int64
}

func (d *Decrypt) Decrypt(inFile string, outFile string, key string, bufferSize int64) error {

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

	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	iv := make([]byte, b.BlockSize())
	msgLen := fi.Size() - int64(len(iv))
	_, err = f.ReadAt(iv, msgLen)
	if err != nil {
		log.Fatal(err)
	}

	outfile, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	log.Printf("Decrypting backup...")

	// The buffer size must be multiple of 16 bytes
	buf := make([]byte, 1024)
	stream := cipher.NewCTR(b, iv)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			// The last bytes are the IV, don't belong the original message
			if n > int(msgLen) {
				n = int(msgLen)
			}
			msgLen -= int64(n)
			stream.XORKeyStream(buf, buf[:n])
			// Write into file
			outfile.Write(buf[:n])
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read %d bytes: %v", n, err)
			break
		}
	}

	err = os.Remove(inFile)
	if err != nil {
		return err
	}

	return err
}
