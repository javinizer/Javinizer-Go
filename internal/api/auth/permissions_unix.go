//go:build !windows

package auth

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

const credentialFileMode = 0600

func enforceCredentialFilePermissions(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("credential path %s must not be a symlink", path)
	}
	if info.IsDir() {
		return fmt.Errorf("credential path %s is a directory", path)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("credential path %s is not a regular file", path)
	}

	if info.Mode().Perm() == credentialFileMode {
		return nil
	}

	if err := os.Chmod(path, credentialFileMode); err != nil {
		if isUnsupportedPermissionMutation(err) {
			return fmt.Errorf(
				"credential file mode is %o and filesystem does not support chmod to %o: %w",
				info.Mode().Perm(),
				credentialFileMode,
				err,
			)
		}
		return err
	}

	info, err = os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("credential path %s must not be a symlink", path)
	}
	if info.Mode().Perm() != credentialFileMode {
		return fmt.Errorf("credential file mode is %o, expected %o", info.Mode().Perm(), credentialFileMode)
	}
	return nil
}

func isUnsupportedPermissionMutation(err error) bool {
	return errors.Is(err, syscall.EOPNOTSUPP) ||
		errors.Is(err, syscall.ENOTSUP) ||
		errors.Is(err, syscall.EROFS)
}
