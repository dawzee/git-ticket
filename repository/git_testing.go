package repository

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

// This is intended for testing only

func CreateTestRepo(bare bool) TestedRepo {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}

	var creator func(string) (*GitRepo, error)

	if bare {
		creator = InitBareGitRepo
	} else {
		creator = InitGitRepo
	}

	repo, err := creator(dir)
	if err != nil {
		log.Fatal(err)
	}

	config := repo.LocalConfig()
	if err := config.StoreString("user.name", "testuser"); err != nil {
		log.Fatal("failed to set user.name for test repository: ", err)
	}
	if err := config.StoreString("user.email", "testuser@example.com"); err != nil {
		log.Fatal("failed to set user.email for test repository: ", err)
	}

	if err = setupSigningKey(repo); err != nil {
		log.Fatal("failed to set up the sigining key: ", err)
	}

	return repo
}

// setupSigningKey creates a GPG key and sets up the local config so it's used.
// The key id is set as "user.signingkey". For the key to be found, a `gpg`
// wrapper which uses only a custom keyring is created and set as "gpg.program".
// Finally "commit.gpgsign" is set to true so the signing takes place.
func setupSigningKey(repo *GitRepo) error {
	config := repo.LocalConfig()

	// Generate a key pair for signing commits.
	entity, err := openpgp.NewEntity("First Last", "", "fl@example.org", nil)
	if err != nil {
		return fmt.Errorf("failed to create a gpg key pair: %s", err)
	}

	if err = config.StoreString("user.signingkey", entity.PrivateKey.KeyIdString()); err != nil {
		return fmt.Errorf("failed to store the key id in git configuration: %s", err)
	}

	// Armor the private part.
	privBuilder := &strings.Builder{}
	w, err := armor.Encode(privBuilder, openpgp.PrivateKeyType, nil)
	if err != nil {
		return fmt.Errorf("failed to armor the private part of the gpg key: %s", err)
	}

	if err = entity.SerializePrivate(w, nil); err != nil {
		return fmt.Errorf("failed to serialize the private key: %s", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close the serialized private key: %s", err)
	}

	armoredPriv := privBuilder.String()

	// Armor the public part.
	pubBuilder := &strings.Builder{}
	if w, err = armor.Encode(pubBuilder, openpgp.PublicKeyType, nil); err != nil {
		return fmt.Errorf("failed to armor the public part of the gpg key: %s", err)
	}

	if err = entity.Serialize(w); err != nil {
		return fmt.Errorf("failed to serialize the public key: %s", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close the serialized public key: %s", err)
	}

	armoredPub := pubBuilder.String()

	// Create a custom gpg keyring to be used when creating commits with `git`.
	keyring, err := ioutil.TempFile("", "keyring")
	if err != nil {
		return fmt.Errorf("failed to create a tempfile for the keyring: %s", err)
	}

	// Import the armored private key to the custom keyring.
	priv, err := ioutil.TempFile("", "privkey")
	if err != nil {
		return fmt.Errorf("failed to create a tempfile for the private key: %s", err)
	}

	if _, err = fmt.Fprint(priv, armoredPriv); err != nil {
		return fmt.Errorf("failed to write the private key: %s", err)
	}

	if err = priv.Close(); err != nil {
		return fmt.Errorf("failed to close the private key file: %s", err)
	}

	if err = exec.Command("gpg", "--no-default-keyring", "--keyring", keyring.Name(), "--import", priv.Name()).Run(); err != nil {
		return fmt.Errorf("failed to import the private key: %s", err)
	}

	// Import the armored public key to the custom keyring.
	pub, err := ioutil.TempFile("", "pubkey")
	if err != nil {
		return fmt.Errorf("failed to create a tempfile for the public key: %s", err)
	}

	if _, err = fmt.Fprint(pub, armoredPub); err != nil {
		return fmt.Errorf("failed to write the public key: %s", err)
	}

	if err = pub.Close(); err != nil {
		return fmt.Errorf("failed to close the public key file: %s", err)
	}

	if err = exec.Command("gpg", "--no-default-keyring", "--keyring", keyring.Name(), "--import", pub.Name()).Run(); err != nil {
		return fmt.Errorf("failed to import the public key: %s", err)
	}

	// Use a gpg wrapper to use a custom keyring containing GPGKeyID.
	gpgWrapper, err := createGPGWrapper(keyring.Name())
	if err != nil {
		return fmt.Errorf("failed to create the GPG wrapper program: %s", err)
	}

	if err := config.StoreString("gpg.program", gpgWrapper); err != nil {
		return fmt.Errorf("failed to set gpg.program for test repository: %s", err)
	}

	if err := config.StoreString("commit.gpgsign", "true"); err != nil {
		return fmt.Errorf("failed to set commit.gpgsign for test repository: %s", err)
	}
	return nil
}

// createGPGWrapper creates a shell script running gpg with a specific keyring.
func createGPGWrapper(keyringPath string) (string, error) {
	file, err := ioutil.TempFile("", "gpgwrapper")
	if err != nil {
		return "", fmt.Errorf("failed to open the tempfile for gpg wrapper: %s", err)
	}

	_, err = fmt.Fprintf(file, `#!/bin/sh
exec gpg --no-default-keyring --keyring="%s" "$@"
`, keyringPath)
	if err != nil {
		return "", fmt.Errorf("failed to write the gpg wrapper: %s", err)
	}

	if err = file.Close(); err != nil {
		return "", fmt.Errorf("failed to close the gpg wrapper file: %s", err)
	}

	if err = os.Chmod(file.Name(), os.FileMode(0700)); err != nil {
		return "", fmt.Errorf("failed to change the file permissions of the gpg wrapper: %s", err)
	}

	return file.Name(), nil
}

func SetupReposAndRemote() (repoA, repoB, remote TestedRepo) {
	repoA = CreateTestRepo(false)
	repoB = CreateTestRepo(false)
	remote = CreateTestRepo(true)

	remoteAddr := "file://" + remote.GetPath()

	err := repoA.AddRemote("origin", remoteAddr)
	if err != nil {
		log.Fatal(err)
	}

	err = repoB.AddRemote("origin", remoteAddr)
	if err != nil {
		log.Fatal(err)
	}

	return repoA, repoB, remote
}
