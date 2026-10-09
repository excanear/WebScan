package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// flexStrings aceita tanto uma string quanto um array de strings no JSON
// (alguns arquivos de assinaturas usam "body" como string única).
type flexStrings []string

func (f *flexStrings) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if b[0] == '[' {
		var arr []string
		if err := json.Unmarshal(b, &arr); err != nil {
			return err
		}
		*f = arr
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*f = []string{s}
	return nil
}

// Signature describes a simple fingerprint rule set loaded from JSON.
// Aceita tanto header_contains/body_contains quanto os aliases
// match_headers/match_body usados em alguns arquivos de assinaturas.
type Signature struct {
	Name           string            `json:"name"`
	HeaderContains map[string]string `json:"header_contains,omitempty"`
	BodyContains   flexStrings       `json:"body_contains,omitempty"`
	MatchHeaders   map[string]string `json:"match_headers,omitempty"`
	MatchBody      flexStrings       `json:"match_body,omitempty"`
}

var signatureDB []Signature = []Signature{
	{Name: "WordPress", BodyContains: []string{"wp-content", "wp-includes"}},
	{Name: "Joomla", BodyContains: []string{"content=\"Joomla"}},
	{Name: "Cloudflare", HeaderContains: map[string]string{"Server": "cloudflare"}},
}

// LoadSignatures attempts to load additional signatures from a JSON file.
// If the file cannot be read, the embedded defaults remain in use.
func LoadSignatures(path string) error {
	// If no path provided, try common default locations.
	if path == "" {
		defaults := []string{
			"signatures.json",
			"test/signatures/popular_signatures.json",
			"test/signatures/expanded_signatures.json",
			"test/signatures/sample_signatures.json",
		}
		for _, p := range defaults {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
		if path == "" {
			// nothing to load
			return nil
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	var s []Signature
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	// append to in-memory DB
	signatureDB = append(signatureDB, s...)
	return nil
}

// MatchSignatures returns names of matching signatures for given headers/body.
func MatchSignatures(headers http.Header, body []byte) []string {
	var found []string
	bstr := strings.ToLower(string(body))
	for _, sig := range signatureDB {
		// combina os campos canônicos com os aliases match_headers/match_body
		hdrChecks := map[string]string{}
		for k, v := range sig.HeaderContains {
			hdrChecks[k] = v
		}
		for k, v := range sig.MatchHeaders {
			hdrChecks[k] = v
		}
		bodyChecks := append(append([]string{}, sig.BodyContains...), sig.MatchBody...)

		// uma assinatura sem NENHUM critério não deve casar (evita falso-positivo
		// quando o JSON usa uma chave desconhecida e os campos ficam vazios).
		if len(hdrChecks) == 0 && len(bodyChecks) == 0 {
			continue
		}

		ok := true
		for hk, hv := range hdrChecks {
			if hv == "" {
				// valor vazio = critério de PRESENÇA: o header precisa existir
				// (senão strings.Contains(x, "") casaria com qualquer resposta).
				if len(headers.Values(hk)) == 0 {
					ok = false
					break
				}
				continue
			}
			if !strings.Contains(strings.ToLower(headers.Get(hk)), strings.ToLower(hv)) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for _, bc := range bodyChecks {
			if !strings.Contains(bstr, strings.ToLower(bc)) {
				ok = false
				break
			}
		}
		if ok {
			found = append(found, sig.Name)
		}
	}
	return found
}
