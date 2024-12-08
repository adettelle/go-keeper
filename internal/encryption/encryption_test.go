package encryption

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryption(t *testing.T) {
	plainText := "hello world --------------------- asdasdasd ---------------- asdadas very long text"

	passphrase := "my_super_secret_password1"
	encryptedString, err := AESEncrypt(plainText, []byte(passphrase))
	require.NoError(t, err)

	decryptedString, err := AESDecrypt(encryptedString, []byte(passphrase))
	require.NoError(t, err)

	require.Equal(t, decryptedString, plainText)
}

func TestEncryptionKeyTooShort(t *testing.T) {
	plainText := "hello world --------------------- asdasdasd ---------------- asdadas very long text"

	passphrase := "short_key"
	_, err := AESEncrypt(plainText, []byte(passphrase))
	require.Error(t, err)
}

func TestDecryptionKeyTooShort(t *testing.T) {
	plainText := "hello world --------------------- asdasdasd ---------------- asdadas very long text"

	passphrase := "short_key"
	_, err := AESDecrypt(plainText, []byte(passphrase))
	require.Error(t, err)
}

func TestPrepareKey(t *testing.T) {
	key := "123456781234567890"
	preparedKey, err := prepareKey([]byte(key))
	require.NoError(t, err)
	require.Len(t, preparedKey, 16)
}

func TestPrepareKeyTooShort(t *testing.T) {
	key := "123"
	_, err := prepareKey([]byte(key))
	require.Error(t, err)
}
