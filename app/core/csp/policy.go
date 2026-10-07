package csp

import (
	"slices"
	"strings"
)

type Service struct {
	ScriptSrc []string

	FrameSrc []string

	TrustedTypes []string
}

type Policy struct {
	Services []Service
}

type Header struct {
	scriptSrc string

	rest string
}

func (p Policy) Header() Header {
	directives := []string{
		"img-src 'self' data: https:",
		"font-src 'self'",
		"connect-src 'self' https:",
	}
	if frames := p.sources(frameSrc); frames != "" {
		directives = append(directives, "frame-src "+frames)
	}

	directives = append(directives, "manifest-src 'self'", "form-action 'self'", "frame-ancestors 'none'", "base-uri 'none'", "require-trusted-types-for 'script'", "trusted-types "+p.trustedTypes())

	return Header{scriptSrc: p.sources(scriptSrc, "'strict-dynamic'"), rest: strings.Join(directives, "; ")}
}

func (h Header) Value(nonce string) string {
	source := "'nonce-" + nonce + "'"

	return "default-src 'none'; script-src " + source + " " + h.scriptSrc + "; style-src 'self' " + source + "; " + h.rest
}

func (p Policy) sources(pick func(*Service) []string, fixed ...string) string {
	all := slices.Clone(fixed)

	for i := range p.Services {
		for _, source := range pick(&p.Services[i]) {
			if slices.Contains(all, source) == false {
				all = append(all, source)
			}
		}
	}

	return strings.Join(all, " ")
}

func (p Policy) trustedTypes() string {
	if names := p.sources(trustedTypes); names != "" {
		return names
	}

	return "'none'"
}

func scriptSrc(s *Service) []string { return s.ScriptSrc }

func frameSrc(s *Service) []string { return s.FrameSrc }

func trustedTypes(s *Service) []string { return s.TrustedTypes }
