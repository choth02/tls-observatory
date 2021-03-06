package certificate

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	Ubuntu_TS_name    = "Ubuntu"
	Mozilla_TS_name   = "Mozilla"
	Microsoft_TS_name = "Microsoft"
	Apple_TS_name     = "Apple"
	Android_TS_name   = "Android"
)

type Certificate struct {
	ID                     int64                     `json:"id"`
	Serial                 string                    `json:"serialNumber,omitempty"`
	ScanTarget             string                    `json:"scanTarget,omitempty"`
	IPs                    []string                  `json:"ips,omitempty"`
	Version                int                       `json:"version,omitempty"`
	SignatureAlgorithm     string                    `json:"signatureAlgorithm,omitempty"`
	Issuer                 Subject                   `json:"issuer,omitempty"`
	Validity               Validity                  `json:"validity,omitempty"`
	Subject                Subject                   `json:"subject,omitempty"`
	Key                    SubjectPublicKeyInfo      `json:"key,omitempty"`
	X509v3Extensions       Extensions                `json:"x509v3Extensions,omitempty"`
	X509v3BasicConstraints string                    `json:"x509v3BasicConstraints,omitempty"`
	CA                     bool                      `json:"ca,omitempty"`
	Analysis               interface{}               `json:"analysis,omitempty"` //for future use...
	ParentSignature        []string                  `json:"parentSignature,omitempty"`
	ValidationInfo         map[string]ValidationInfo `json:"validationInfo,omitempty"`
	FirstSeenTimestamp     time.Time                 `json:"firstSeenTimestamp"`
	LastSeenTimestamp      time.Time                 `json:"lastSeenTimestamp"`
	Hashes                 Hashes                    `json:"hashes,omitempty"`
	Raw                    string                    `json:"Raw,omitempty"`
	Anomalies              string                    `json:"anomalies,omitempty"`
}

type Hashes struct {
	MD5               string `json:"md5,omitempty"`
	SHA1              string `json:"sha1,omitempty"`
	SHA256            string `json:"sha256,omitempty"`
	SHA256SubjectSPKI string `json:"sha256_subject_spki,omitempty"`
	PKPSHA256         string `json:"pin-sha256,omitempty"`
}

type Validity struct {
	NotBefore time.Time `json:"notBefore"`
	NotAfter  time.Time `json:"notAfter"`
}

type Subject struct {
	ID           int64    `json:"id,omitempty"`
	Country      []string `json:"c,omitempty"`
	Organisation []string `json:"o,omitempty"`
	OrgUnit      []string `json:"ou,omitempty"`
	CommonName   string   `json:"cn,omitempty"`
}

type SubjectPublicKeyInfo struct {
	Alg      string  `json:"alg,omitempty"`
	Size     float64 `json:"size,omitempty"`
	Exponent float64 `json:"exponent,omitempty"`
	X        string  `json:"x,omitempty"`
	Y        string  `json:"y,omitempty"`
	P        string  `json:"p,omitempty"`
	Q        string  `json:"q,omitempty"`
	G        string  `json:"g,omitempty"`
	Curve    string  `json:"curve,omitempty"`
}

//Currently exporting extensions that are already decoded into the x509 Certificate structure
type Extensions struct {
	AuthorityKeyId         string   `json:"authorityKeyId,omitempty"`
	SubjectKeyId           string   `json:"subjectKeyId,omitempty"`
	KeyUsage               []string `json:"keyUsage,omitempty"`
	ExtendedKeyUsage       []string `json:"extendedKeyUsage,omitempty"`
	SubjectAlternativeName []string `json:"subjectAlternativeName,omitempty"`
	CRLDistributionPoints  []string `json:"crlDistributionPoint,omitempty"`
	PolicyIdentifiers      []string `json:"policyIdentifiers,omitempty"`
	IsNameConstrained      bool     `json:"isNameConstrained,omitempty"`
	PermittedNames         []string `json:"permittedNames,omitempty"`
}

type X509v3BasicConstraints struct {
	CA       bool        `json:"ca,omitempty"`
	Analysis interface{} `json:"analysis,omitempty"`
}

type Chain struct {
	Domain string   `json:"domain"`
	IP     string   `json:"ip"`
	Certs  []string `json:"certs"`
}

type IDs struct {
	_type  string   `json:"type"`
	values []string `json:"values"`
}

type JsonRawCert struct {
	RawCert string `json:"rawCert"`
}

type TrustStore struct {
	Name  string
	Certs *x509.CertPool
}

type ValidationInfo struct {
	IsValid         bool   `json:"isValid,omitempty"`
	ValidationError string `json:"validationError,omitempty"`
}

type Trust struct {
	ID               int64
	CertID           int64
	IssuerID         int64
	Timestamp        time.Time
	TrustUbuntu      bool
	TrustMozilla     bool
	TrustedMicrosoft bool
	TrustedApple     bool
	TrustedAndroid   bool
	Current          bool
}

var SignatureAlgorithm = [...]string{
	"UnknownSignatureAlgorithm",
	"MD2WithRSA",
	"MD5WithRSA",
	"SHA1WithRSA",
	"SHA256WithRSA",
	"SHA384WithRSA",
	"SHA512WithRSA",
	"DSAWithSHA1",
	"DSAWithSHA256",
	"ECDSAWithSHA1",
	"ECDSAWithSHA256",
	"ECDSAWithSHA384",
	"ECDSAWithSHA512",
}

var ExtKeyUsage = [...]string{
	"ExtKeyUsageAny",
	"ExtKeyUsageServerAuth",
	"ExtKeyUsageClientAuth",
	"ExtKeyUsageCodeSigning",
	"ExtKeyUsageEmailProtection",
	"ExtKeyUsageIPSECEndSystem",
	"ExtKeyUsageIPSECTunnel",
	"ExtKeyUsageIPSECUser",
	"ExtKeyUsageTimeStamping",
	"ExtKeyUsageOCSPSigning",
	"ExtKeyUsageMicrosoftServerGatedCrypto",
	"ExtKeyUsageNetscapeServerGatedCrypto",
}

var PublicKeyAlgorithm = [...]string{
	"UnknownPublicKeyAlgorithm",
	"RSA",
	"DSA",
	"ECDSA",
}

func SHA256SubjectSPKI(cert *x509.Certificate) string {
	h := sha256.New()
	h.Write(cert.RawSubject)
	h.Write(cert.RawSubjectPublicKeyInfo)
	return fmt.Sprintf("%X", h.Sum(nil))
}

func PKPSHA256Hash(cert *x509.Certificate) string {
	h := sha256.New()
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		der, _ := x509.MarshalPKIXPublicKey(pub)
		h.Write(der)
	case *dsa.PublicKey:
		der, _ := x509.MarshalPKIXPublicKey(pub)
		h.Write(der)
	case *ecdsa.PublicKey:
		der, _ := x509.MarshalPKIXPublicKey(pub)
		h.Write(der)
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func SHA256Hash(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%X", h[:])
}

func MD5Hash(data []byte) string {
	h := md5.Sum(data)
	return fmt.Sprintf("%X", h[:])
}

func SHA1Hash(data []byte) string {
	h := sha1.Sum(data)
	return fmt.Sprintf("%X", h[:])
}

//GetBooleanValidity converts the validation info map to DB booleans
func (c Certificate) GetBooleanValidity() (trusted_ubuntu, trusted_mozilla, trusted_microsoft, trusted_apple, trusted_android bool) {

	//check Ubuntu validation info
	valInfo, ok := c.ValidationInfo[Ubuntu_TS_name]

	if !ok {
		trusted_ubuntu = false
	} else {
		trusted_ubuntu = valInfo.IsValid
	}

	//check Mozilla validation info
	valInfo, ok = c.ValidationInfo[Mozilla_TS_name]

	if !ok {
		trusted_mozilla = false
	} else {
		trusted_mozilla = valInfo.IsValid
	}

	//check Microsoft validation info
	valInfo, ok = c.ValidationInfo[Microsoft_TS_name]

	if !ok {
		trusted_microsoft = false
	} else {
		trusted_microsoft = valInfo.IsValid
	}

	//check Apple validation info
	valInfo, ok = c.ValidationInfo[Apple_TS_name]

	if !ok {
		trusted_apple = false
	} else {
		trusted_apple = valInfo.IsValid
	}

	//check Android validation info
	valInfo, ok = c.ValidationInfo[Android_TS_name]

	if !ok {
		trusted_android = false
	} else {
		trusted_android = valInfo.IsValid
	}
	return
}

// GetValidityMap converts boolean validity variables to a validity map.
func GetValidityMap(trusted_ubuntu, trusted_mozilla, trusted_microsoft, trusted_apple, trusted_android bool) map[string]ValidationInfo {

	vUbuntu := ValidationInfo{IsValid: trusted_ubuntu}
	vMozilla := ValidationInfo{IsValid: trusted_mozilla}
	vMicrosoft := ValidationInfo{IsValid: trusted_microsoft}
	vApple := ValidationInfo{IsValid: trusted_apple}
	vAndroid := ValidationInfo{IsValid: trusted_android}

	m := make(map[string]ValidationInfo)

	m[Ubuntu_TS_name] = vUbuntu
	m[Mozilla_TS_name] = vMozilla
	m[Microsoft_TS_name] = vMicrosoft
	m[Apple_TS_name] = vApple
	m[Android_TS_name] = vAndroid

	return m

}

func getExtKeyUsages(cert *x509.Certificate) []string {
	usage := make([]string, 0)
	for _, eku := range cert.ExtKeyUsage {
		usage = append(usage, ExtKeyUsage[eku])
	}
	for _, unknownEku := range cert.UnknownExtKeyUsage {
		usage = append(usage, unknownEku.String())
	}
	return usage
}

func getPolicyIdentifiers(cert *x509.Certificate) []string {
	identifiers := make([]string, 0)
	for _, pi := range cert.PolicyIdentifiers {
		identifiers = append(identifiers, pi.String())
	}
	return identifiers
}

func getKeyUsages(cert *x509.Certificate) []string {
	usage := make([]string, 0)
	keyUsage := cert.KeyUsage

	//calculate included keyUsage from bitmap
	//String values taken from OpenSSL

	if keyUsage&x509.KeyUsageDigitalSignature != 0 {
		usage = append(usage, "Digital Signature")
	}
	if keyUsage&x509.KeyUsageContentCommitment != 0 {
		usage = append(usage, "Non Repudiation")
	}

	if keyUsage&x509.KeyUsageKeyEncipherment != 0 {
		usage = append(usage, "Key Encipherment")
	}

	if keyUsage&x509.KeyUsageDataEncipherment != 0 {
		usage = append(usage, "Data Encipherment")
	}

	if keyUsage&x509.KeyUsageKeyAgreement != 0 {
		usage = append(usage, "Key Agreement")
	}

	if keyUsage&x509.KeyUsageCertSign != 0 {
		usage = append(usage, "Certificate Sign")
	}

	if keyUsage&x509.KeyUsageCRLSign != 0 {
		usage = append(usage, "CRL Sign")
	}

	if keyUsage&x509.KeyUsageEncipherOnly != 0 {
		usage = append(usage, "Encipher Only")
	}

	if keyUsage&x509.KeyUsageDecipherOnly != 0 {
		usage = append(usage, "Decipher Only")
	}

	return usage
}

//getCertExtensions currently stores only the extensions that are already exported by GoLang
//(in the x509 Certificate Struct)
func getCertExtensions(cert *x509.Certificate) Extensions {
	// initialize []string to store them as `[]` instead of null
	san := make([]string, 0)
	san = append(san, cert.DNSNames...)
	crld := make([]string, 0)
	crld = append(crld, cert.CRLDistributionPoints...)
	pnames := make([]string, 0)
	pnames = append(pnames, cert.PermittedDNSDomains...)
	ext := Extensions{
		AuthorityKeyId:         base64.StdEncoding.EncodeToString(cert.AuthorityKeyId),
		SubjectKeyId:           base64.StdEncoding.EncodeToString(cert.SubjectKeyId),
		KeyUsage:               getKeyUsages(cert),
		ExtendedKeyUsage:       getExtKeyUsages(cert),
		PolicyIdentifiers:      getPolicyIdentifiers(cert),
		SubjectAlternativeName: san,
		CRLDistributionPoints:  crld,
		PermittedNames:         pnames,
	}
	if len(ext.PermittedNames) > 0 {
		ext.IsNameConstrained = true
	}
	return ext
}

func getPublicKeyInfo(cert *x509.Certificate) (SubjectPublicKeyInfo, error) {
	pubInfo := SubjectPublicKeyInfo{
		Alg: PublicKeyAlgorithm[cert.PublicKeyAlgorithm],
	}

	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		pubInfo.Size = float64(pub.N.BitLen())
		pubInfo.Exponent = float64(pub.E)

	case *dsa.PublicKey:
		pubInfo.Size = float64(pub.Y.BitLen())
		textInt, err := pub.G.MarshalText()

		if err == nil {
			pubInfo.G = string(textInt)
		} else {
			return pubInfo, err
		}

		textInt, err = pub.P.MarshalText()

		if err == nil {
			pubInfo.P = string(textInt)
		} else {
			return pubInfo, err
		}

		textInt, err = pub.Q.MarshalText()

		if err == nil {
			pubInfo.Q = string(textInt)
		} else {
			return pubInfo, err
		}

		textInt, err = pub.Y.MarshalText()

		if err == nil {
			pubInfo.Y = string(textInt)
		} else {
			return pubInfo, err
		}

	case *ecdsa.PublicKey:
		pubInfo.Size = float64(pub.Curve.Params().BitSize)
		pubInfo.Curve = pub.Curve.Params().Name
		pubInfo.Y = pub.Y.String()
		pubInfo.X = pub.X.String()
	}

	return pubInfo, nil

}

//certtoStored returns a Certificate struct created from a X509.Certificate
func CertToStored(cert *x509.Certificate, parentSignature, domain, ip string, TSName string, valInfo *ValidationInfo) Certificate {
	var (
		err    error
		stored = Certificate{}
	)
	// initialize []string to never store them as null
	stored.ParentSignature = make([]string, 0)
	stored.IPs = make([]string, 0)

	stored.Version = cert.Version
	stored.Serial = strings.ToUpper(hex.EncodeToString(cert.SerialNumber.Bytes()))
	stored.SignatureAlgorithm = SignatureAlgorithm[cert.SignatureAlgorithm]

	stored.Key, err = getPublicKeyInfo(cert)
	if err != nil {
		log.Printf("Failed to retrieve public key information: %v. Continuing anyway.", err)
	}

	stored.Issuer.Country = cert.Issuer.Country
	stored.Issuer.Organisation = cert.Issuer.Organization
	stored.Issuer.OrgUnit = cert.Issuer.OrganizationalUnit
	stored.Issuer.CommonName = cert.Issuer.CommonName

	stored.Subject.Country = cert.Subject.Country
	stored.Subject.Organisation = cert.Subject.Organization
	stored.Subject.OrgUnit = cert.Subject.OrganizationalUnit
	stored.Subject.CommonName = cert.Subject.CommonName

	stored.Validity.NotBefore = cert.NotBefore.UTC()
	stored.Validity.NotAfter = cert.NotAfter.UTC()

	stored.X509v3Extensions = getCertExtensions(cert)

	//below check tries to hack around the basic constraints extension
	//not being available in versions < 3.
	//Only the IsCa variable is set, as setting X509v3BasicConstraints
	//messes up the validation procedure.
	if cert.Version < 3 {
		stored.CA = cert.IsCA
	} else {
		if cert.BasicConstraintsValid {
			stored.X509v3BasicConstraints = "Critical"
			stored.CA = cert.IsCA
		} else {
			stored.X509v3BasicConstraints = ""
			stored.CA = false
		}
	}

	t := time.Now().UTC()

	stored.FirstSeenTimestamp = t
	stored.LastSeenTimestamp = t

	stored.ParentSignature = append(stored.ParentSignature, parentSignature)

	if !cert.IsCA {
		stored.ScanTarget = domain
		stored.IPs = append(stored.IPs, ip)
	}

	stored.ValidationInfo = make(map[string]ValidationInfo)
	stored.ValidationInfo[TSName] = *valInfo

	stored.Hashes.MD5 = MD5Hash(cert.Raw)
	stored.Hashes.SHA1 = SHA1Hash(cert.Raw)
	stored.Hashes.SHA256 = SHA256Hash(cert.Raw)
	stored.Hashes.SHA256SubjectSPKI = SHA256SubjectSPKI(cert)
	stored.Hashes.PKPSHA256 = PKPSHA256Hash(cert)

	stored.Raw = base64.StdEncoding.EncodeToString(cert.Raw)

	return stored
}

// ToX509() returns the crypto/x509 version of a certificate
func (cert Certificate) ToX509() (xcert *x509.Certificate, err error) {
	certRaw, err := base64.StdEncoding.DecodeString(cert.Raw)
	if err != nil {
		return
	}
	return x509.ParseCertificate(certRaw)
}

//printRawCertExtensions Print raw extension info
//for debugging purposes
func printRawCertExtensions(cert *x509.Certificate) {

	for i, extension := range cert.Extensions {

		var numbers string
		for num, num2 := range extension.Id {

			numbers = numbers + " " + "[" + strconv.Itoa(num) + " " + strconv.Itoa(num2) + "]"

		}
		fmt.Println("//", strconv.Itoa(i), ": {", numbers, "}", string(extension.Value))
	}

}

// String() prints the subject as a single string, following OpenSSL's display
// format: Subject: C=US, ST=California, L=Mountain View, O=Google Inc, CN=*.google.com
func (s Subject) String() string {
	var comp []string
	if len(s.Country) > 0 {
		comp = append(comp, "C="+strings.Join(s.Country, ", C="))
	}
	if len(s.Organisation) > 0 {
		comp = append(comp, "O="+strings.Join(s.Organisation, ", O="))
	}
	if len(s.OrgUnit) > 0 {
		comp = append(comp, "OU="+strings.Join(s.OrgUnit, ", OU="))
	}
	if len(s.CommonName) > 0 {
		comp = append(comp, "CN="+s.CommonName)
	}
	return strings.Join(comp, ", ")
}

// IsSelfSigned return true if the subject and issuer fields of a certificate
// are identical
func (c Certificate) IsSelfSigned() bool {
	if c.Subject.CommonName != c.Issuer.CommonName ||
		len(c.Subject.Organisation) != len(c.Issuer.Organisation) ||
		len(c.Subject.OrgUnit) != len(c.Issuer.OrgUnit) ||
		len(c.Subject.Country) != len(c.Issuer.Country) {
		return false
	}
	for i, _ := range c.Subject.Organisation {
		if c.Subject.Organisation[i] != c.Issuer.Organisation[i] {
			return false
		}
	}
	for i, _ := range c.Subject.OrgUnit {
		if c.Subject.OrgUnit[i] != c.Issuer.OrgUnit[i] {
			return false
		}
	}
	for i, _ := range c.Subject.Country {
		if c.Subject.Country[i] != c.Issuer.Country[i] {
			return false
		}
	}
	return true
}
