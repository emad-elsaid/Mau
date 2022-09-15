package main

import (
	"bytes"
	"path"
	"testing"
)

func TestAddFriend(t *testing.T) {
	dir := t.TempDir()
	account, _ := NewAccount(dir, "Ahmed Mohamed", "ahmed@example.com", "strong password")

	friend_dir := t.TempDir()
	friend_account, _ := NewAccount(friend_dir, "Mohamed Mahmoud", "mohamed@example.com", "strong password")
	fingerprint := friend_account.Fingerprint()
	friend_account_pub, _ := friend_account.Export()
	friend, err := AddFriend(account, bytes.NewBuffer(friend_account_pub))
	ASSERT_ERROR(t, nil, err)
	ASSERT_FILE_EXISTS(t, path.Join(dir, ".mau", fingerprint+".pgp"))

	t.Run("Email", func(t T) {
		ASSERT_EQUAL(t, "mohamed@example.com", friend.Email())
	})

	t.Run("Name", func(t T) {
		ASSERT_EQUAL(t, "Mohamed Mahmoud", friend.Name())
	})

	t.Run("Fingerprint", func(t T) {
		ASSERT_EQUAL(t, friend_account.Fingerprint(), friend.Fingerprint())
	})

	t.Run("Identity", func(t T) {
		friend_account_identity, _ := friend_account.Identity()
		friend_identity, err := friend.Identity()
		ASSERT_ERROR(t, nil, err)
		ASSERT_EQUAL(t, friend_account_identity, friend_identity)
	})
}

func TestRemoveFriend(t *testing.T) {
	dir := t.TempDir()
	account, _ := NewAccount(dir, "Ahmed Mohamed", "ahmed@example.com", "strong password")

	friend_dir := t.TempDir()
	friend_account, _ := NewAccount(friend_dir, "Mohamed Mahmoud", "mohamed@example.com", "strong password")
	friend_account_pub, _ := friend_account.Export()
	fingerprint := friend_account.Fingerprint()
	friend, _ := AddFriend(account, bytes.NewBuffer(friend_account_pub))

	err := RemoveFriend(account, friend)
	ASSERT_ERROR(t, nil, err)
	REFUTE_FILE_EXISTS(t, path.Join(dir, ".mau", fingerprint+".pgp"))
}
