package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/argon2"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run main.go <password>")
	}
	password := os.Args[1]
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		log.Fatal(err)
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	fmt.Printf("$argon2id$v=%d$m=65536,t=3,p=4$%s$%s\n",
		argon2.Version,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}
