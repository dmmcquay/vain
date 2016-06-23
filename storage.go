package vain

import "time"

// Storer defines the db interface.
type Storer interface {
	AddPackage(p Package) error
	Confirm(token string) (string, error)
	NSForToken(ns string, tok string) error
	Package(path string) (Package, error)
	PackageExists(path string) bool
	Pkgs() []Package
	Register(email string) (string, error)
	RemovePackage(path string) error
	forgot(email string, window time.Duration) (string, error)
}
