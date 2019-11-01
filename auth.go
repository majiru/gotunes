package main

import (
	auth "gopkg.in/korylprince/go-ad-auth.v2"
)

type UserLevel int

const (
	Basic UserLevel = iota
	Admin
)

func getLevel(groups []string) UserLevel {
	lookup := make(map[string]struct{})
	for _, u := range conf.Admins {
		lookup[u] = struct{}{}
	}
	for _, g := range groups {
		if _, ok := lookup[g]; ok {
			return Admin
		}
	}
	return Basic
}

func AuthUser(username, password string) (bool, UserLevel, error) {
	lc := &auth.Config{
		Server: conf.DAServer,
		Port: 389,
		BaseDN: conf.DABaseDN,
	}
	ok, _, groups, err := auth.AuthenticateExtended(lc, username, password, []string{"cn", "memberOf"}, append(conf.Admins, conf.Users...))
	switch {
	case err != nil:
		return false, 0, err
	case ok:
		return true, getLevel(groups), nil
	default:
		return false, 0, nil
	}
}
