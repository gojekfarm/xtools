package xloadtype

import "net/url"

// URL represents a URI reference.
//
// URL is a type alias for url.URL.
// The general form represented is: [scheme:][//[userinfo@]host][/]path[?query][#fragment]
// See https://tools.ietf.org/html/rfc3986
type URL url.URL

func (u *URL) String() string { return (*url.URL)(u).String() }

func (u *URL) Decode(v string) error {
	parsed, err := url.Parse(v)
	if err != nil {
		return err
	}

	*u = URL(*parsed)

	return nil
}

// Endpoint returns the endpoint of the URL.
// The URL host must be in the form of `host:port`.
func (u *URL) Endpoint() (*Endpoint, error) {
	e := new(Endpoint)

	if err := e.Decode(u.Host); err != nil {
		return nil, err
	}

	return e, nil
}
