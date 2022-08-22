package main

// copied from 	https://raw.githubusercontent.com/charmbracelet/soft-serve/main/config/auth.go
// for the one function that compares keys

import (
	"log"
	"strings"
	"sync"

	gm "github.com/charmbracelet/wish/git"
	"github.com/gliderlabs/ssh"
	"honnef.co/go/tools/config"
)

type Config struct {
	Name         string         `yaml:"name"`
	Host         string         `yaml:"host"`
	Port         int            `yaml:"port"`
	AnonAccess   string         `yaml:"anon-access"`
	AllowKeyless bool           `yaml:"allow-keyless"`
	Users        []User         `yaml:"users"`
	Repos        []MenuRepo     `yaml:"repos"`
	Source       *RepoSource    `yaml:"-"`
	Cfg          *config.Config `yaml:"-"`
	mtx          sync.Mutex
}

func (cfg *Config) accessForKey(repo string, pk ssh.PublicKey) gm.AccessLevel {
	private := cfg.isPrivate(repo)
	for _, u := range cfg.Users {
		for _, k := range u.PublicKeys {
			apk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(k)))
			if err != nil {
				log.Printf("error: malformed authorized key: '%s'", k)
				return gm.NoAccess
			}
			if ssh.KeysEqual(pk, apk) {
				if u.Admin {
					return gm.AdminAccess
				}
				for _, r := range u.CollabRepos {
					if repo == r {
						return gm.ReadWriteAccess
					}
				}
				if !private {
					return gm.ReadOnlyAccess
				}
			}
		}
	}
	if private && len(cfg.Users) > 0 {
		return gm.NoAccess
	}
	switch cfg.AnonAccess {
	case "no-access":
		return gm.NoAccess
	case "read-only":
		return gm.ReadOnlyAccess
	case "read-write":
		return gm.ReadWriteAccess
	case "admin-access":
		return gm.AdminAccess
	default:
		return gm.NoAccess
	}
}
