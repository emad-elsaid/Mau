package mau

import (
	"bytes"
	"context"
	"log"
	"net"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	account, err := NewAccount(t.TempDir(), "Ahmed Mohamed", "ahmed@example.com", "strong password")

	client, err := account.Client(account.Fingerprint(), nil)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, client)
}

func TestDownloadFriend(t *testing.T) {
	account_dir := t.TempDir()
	account, _ := NewAccount(account_dir, "Ahmed Mohamed", "ahmed@example.com", "strong password")
	var account_key bytes.Buffer
	account.Export(&account_key)

	friend_dir := t.TempDir()
	friend, _ := NewAccount(friend_dir, "Mohamed Mahmoud", "mohamed@example.com", "strong password")
	var friend_key bytes.Buffer
	friend.Export(&friend_key)
	server, _ := friend.Server(nil)

	listener, address := TempListener()
	go server.Serve(*listener, "")
	defer server.Close()

	client, _ := account.Client(friend.Fingerprint(), nil)

	t.Run("When the fingerprint is not a friend", func(t T) {
		err := client.DownloadFriend(context.Background(), friend.Fingerprint(), time.Now(), []FingerprintResolver{StaticAddress(address)})
		assert.Error(t, ErrFriendNotFollowed, err)
	})

	t.Run("When friend but not followed", func(t T) {
		account.AddFriend(bytes.NewReader(friend_key.Bytes()))
		err := client.DownloadFriend(context.Background(), friend.Fingerprint(), time.Now(), []FingerprintResolver{StaticAddress(address)})
		assert.Error(t, ErrFriendNotFollowed, err)
	})

	t.Run("When friend and followed", func(t T) {
		f, err := account.AddFriend(bytes.NewReader(friend_key.Bytes()))
		assert.NoError(t, err)
		account.Follow(f)

		err = client.DownloadFriend(Timeout(time.Second), friend.Fingerprint(), time.Now(), []FingerprintResolver{StaticAddress(address)})
		assert.NoError(t, err)
	})

	t.Run("When a file is encrypted for friend", func(t T) {
		// Create a file in the friend account
		aFriend, _ := friend.AddFriend(bytes.NewReader(account_key.Bytes()))
		_, err := friend.AddFile(strings.NewReader("Hello world!"), "hello world.txt", []*Friend{aFriend})
		assert.NoError(t, err)
		assert.FileExists(t, path.Join(friend_dir, friend.Fingerprint().String(), "hello world.txt.pgp"))

		// and download it to the account
		err = client.DownloadFriend(Timeout(time.Second), friend.Fingerprint(), time.Now().Add(-time.Second), []FingerprintResolver{StaticAddress(address)})
		assert.NoError(t, err)
		assert.FileExists(t, path.Join(account_dir, friend.Fingerprint().String(), "hello world.txt.pgp"))
	})

	t.Run("When private file exists", func(t T) {
		_, err := friend.AddFile(strings.NewReader("Private social security number"), "private.txt", []*Friend{})
		assert.NoError(t, err)
		assert.FileExists(t, path.Join(friend_dir, friend.Fingerprint().String(), "private.txt.pgp"))

		err = client.DownloadFriend(Timeout(time.Second), friend.Fingerprint(), time.Now().Add(-time.Second), []FingerprintResolver{StaticAddress(address)})
		assert.NoError(t, err)
		assert.NoFileExists(t, path.Join(account_dir, friend.Fingerprint().String(), "private.txt.pgp"))
	})

	t.Run("When no address is provided it find the user on the local network", func(t T) {
		ctx := Timeout(10 * time.Second)
		err := client.DownloadFriend(ctx, friend.Fingerprint(), time.Now().Add(-time.Second), []FingerprintResolver{LocalFriendAddress})
		assert.NoError(t, err)
	})

	t.Run("Connecting to an account with wrong peer ID", func(t T) {
		anotherFriend, _ := NewAccount(t.TempDir(), "Another person", "another@example.com", "password")
		client, _ := account.Client(anotherFriend.Fingerprint(), nil)
		ctx := Timeout(10 * time.Second)
		err := client.DownloadFriend(ctx, friend.Fingerprint(), time.Now().Add(-time.Second), []FingerprintResolver{StaticAddress(address)})
		assert.Error(t, ErrIncorrectPeerCertificate, err)
	})
}

func Timeout(p time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), p)
	return ctx
}

func TempListener() (*net.Listener, string) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		log.Fatal("Error while creating listener for testing:", err.Error())
	}

	address := listener.Addr().(*net.TCPAddr).String()
	url := address
	return &listener, url
}
