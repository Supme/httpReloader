package httpreloader

import (
	"crypto/tls"
	"testing"
)

func TestGetCertificate(t *testing.T) {
	type testData map[string][]string

	goodData := testData{
		"domain.com": []string{
			"domain.com",
			"domain.com.",
			"DomaIn.Com",
			" domain.com. ",
			" domain.com ",
		},
		"*.wildcarddomain.com": []string{
			"www.wildcarddomain.com",
			"WWW.WildcarDdomain.com",
			"www.wildcarddomain.com.",
			" www.wildcarddomain.com ",
			"static.www.wildcarddomain.com",
		},
	}

	badData := testData{
		"domain.com": []string{
			"domain.net",
			"domain.*.com",
			"domain.com.*",
		},
		"*.wildcarddomain.com": []string{
			"static.www.wildcarddomain.net",
			"static.www.wildcarddomain.net.",
			"wildcarddomain.com",
			" wildcarddomain.com ",
		},
	}

	defaultCertData := testData{
		"domain.com": []string{
			"domain.com",
			"default.net",
		},
	}

	check := func(certFile, keyFile string, data testData, ifNil bool) {
		r, _ := NewReloader(certFile, keyFile)
		for k := range data {
			err := r.UpdateCertificate("./test_data/cert1.pem", "./test_data/key1.pem", k)
			if err != nil {
				t.Error(err)
			}
		}
		getCert := r.GetCertificateFunc()
		for k, v := range data {
			for i := range v {
				if ifNil {
					if _, err := getCert(&tls.ClientHelloInfo{ServerName: v[i]}); err == ErrCertificateNotLoaded {
						t.Errorf("certificate %s not found for %s err: %s", k, v[i], err)
					}
				} else {
					if _, err := getCert(&tls.ClientHelloInfo{ServerName: v[i]}); err == nil {
						t.Errorf("certificate %s found for %s", k, v[i])
					}
				}
			}
		}
	}

	check("", "", goodData, true)
	check("", "", badData, false)
	check("./test_data/cert1.pem", "./test_data/key1.pem", defaultCertData, true)
}
