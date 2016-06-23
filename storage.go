package vain

import "time"

// Storer defines the db interface.
type Storer interface {
	NSForToken(ns namespace, tok Token) error

	Package(path string) (Package, error)
	AddPackage(p Package) error
	RemovePackage(pth path) error
	PackageExists(pth path) bool
	Pkgs() []Package

	Register(e Email) (Token, error)
	Confirm(tok Token) (Token, error)
	Forgot(e Email, window time.Duration) (Token, error)
}
