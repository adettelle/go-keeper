package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func MustGenCert(prefix string) {
	// var prefix string

	// flag.StringVar(&prefix, "p", prefix, "certificate prefix")
	// flag.Parse()
	// if prefix == "" {
	// 	log.Fatal("No prefix provided")
	// }

	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1658),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"adettelle"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    []string{"localhost"},
	}

	// создаём новый приватный RSA-ключ длиной 4096 бит
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer

	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		log.Fatal(err)
	}

	var privateKeyPEM bytes.Buffer

	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		log.Fatal(err)
	}

	fileCert, err := os.OpenFile(fmt.Sprintf("./keys/%s_cert.pem", prefix), os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer fileCert.Close()

	_, err = fileCert.Write(certPEM.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	filePrivateKey, err := os.OpenFile(fmt.Sprintf("./keys/%s_privatekey.pem", prefix), os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal("OP", err)
	}
	defer fileCert.Close()

	_, err = filePrivateKey.Write(privateKeyPEM.Bytes())
	if err != nil {
		log.Fatal("/wr", err)
	}

	// savePubKey(&privateKey.PublicKey, prefix)
}

/*
func savePubKey(pubkey *rsa.PublicKey, prefix string) {
	pubASN1, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		log.Fatal(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	file, err := os.OpenFile(fmt.Sprintf("./keys/%s_publickey.pem", prefix), os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write(pubBytes)
	if err != nil {
		log.Fatal(err)
	}
}
*/
