package avatars

import "github.com/James-HoneyBadger/ratel-ledger/pkg/models"

// AvatarProvider is user avatar provider interface
type AvatarProvider interface {
	GetAvatarUrl(user *models.User) string
}
