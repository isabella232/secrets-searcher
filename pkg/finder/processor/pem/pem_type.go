package pem

// *** generated with go-genums ***

// PEMTypeEnum is the the enum interface that can be used
type PEMTypeEnum interface {
	String() string
	Value() string
	uniquePEMTypeMethod()
}

// pEMTypeEnumBase is the internal, non-exported type
type pEMTypeEnumBase struct{ value string }

// Value() returns the enum value
func (eb pEMTypeEnumBase) Value() string { return eb.value }

// String() returns the enum name as you use it in Go code,
// needs to be overriden by inheriting types
func (eb pEMTypeEnumBase) String() string { return "" }

// X509CRL is the enum type for 'valueX509CRL' value
type X509CRL struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueX509CRL'
func (X509CRL) New() PEMTypeEnum { return X509CRL{pEMTypeEnumBase{valueX509CRL}} }

// String returns always "X509CRL" for this enum type
func (X509CRL) String() string { return "X509CRL" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (X509CRL) uniquePEMTypeMethod() {}

// PublicKey is the enum type for 'valuePublicKey' value
type PublicKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valuePublicKey'
func (PublicKey) New() PEMTypeEnum { return PublicKey{pEMTypeEnumBase{valuePublicKey}} }

// String returns always "PublicKey" for this enum type
func (PublicKey) String() string { return "PublicKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (PublicKey) uniquePEMTypeMethod() {}

// EncryptedPrivateKey is the enum type for 'valueEncryptedPrivateKey' value
type EncryptedPrivateKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueEncryptedPrivateKey'
func (EncryptedPrivateKey) New() PEMTypeEnum { return EncryptedPrivateKey{pEMTypeEnumBase{valueEncryptedPrivateKey}} }

// String returns always "EncryptedPrivateKey" for this enum type
func (EncryptedPrivateKey) String() string { return "EncryptedPrivateKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (EncryptedPrivateKey) uniquePEMTypeMethod() {}

// ECParameters is the enum type for 'valueECParameters' value
type ECParameters struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueECParameters'
func (ECParameters) New() PEMTypeEnum { return ECParameters{pEMTypeEnumBase{valueECParameters}} }

// String returns always "ECParameters" for this enum type
func (ECParameters) String() string { return "ECParameters" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (ECParameters) uniquePEMTypeMethod() {}

// Certificate is the enum type for 'valueCertificate' value
type Certificate struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueCertificate'
func (Certificate) New() PEMTypeEnum { return Certificate{pEMTypeEnumBase{valueCertificate}} }

// String returns always "Certificate" for this enum type
func (Certificate) String() string { return "Certificate" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Certificate) uniquePEMTypeMethod() {}

// CertificateRequest is the enum type for 'valueCertificateRequest' value
type CertificateRequest struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueCertificateRequest'
func (CertificateRequest) New() PEMTypeEnum { return CertificateRequest{pEMTypeEnumBase{valueCertificateRequest}} }

// String returns always "CertificateRequest" for this enum type
func (CertificateRequest) String() string { return "CertificateRequest" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (CertificateRequest) uniquePEMTypeMethod() {}

// PGPPrivateKeyBlock is the enum type for 'valuePGPPrivateKeyBlock' value
type PGPPrivateKeyBlock struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valuePGPPrivateKeyBlock'
func (PGPPrivateKeyBlock) New() PEMTypeEnum { return PGPPrivateKeyBlock{pEMTypeEnumBase{valuePGPPrivateKeyBlock}} }

// String returns always "PGPPrivateKeyBlock" for this enum type
func (PGPPrivateKeyBlock) String() string { return "PGPPrivateKeyBlock" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (PGPPrivateKeyBlock) uniquePEMTypeMethod() {}

// AnyPrivateKey is the enum type for 'valueAnyPrivateKey' value
type AnyPrivateKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueAnyPrivateKey'
func (AnyPrivateKey) New() PEMTypeEnum { return AnyPrivateKey{pEMTypeEnumBase{valueAnyPrivateKey}} }

// String returns always "AnyPrivateKey" for this enum type
func (AnyPrivateKey) String() string { return "AnyPrivateKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (AnyPrivateKey) uniquePEMTypeMethod() {}

// RSAPublicKey is the enum type for 'valueRSAPublicKey' value
type RSAPublicKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueRSAPublicKey'
func (RSAPublicKey) New() PEMTypeEnum { return RSAPublicKey{pEMTypeEnumBase{valueRSAPublicKey}} }

// String returns always "RSAPublicKey" for this enum type
func (RSAPublicKey) String() string { return "RSAPublicKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (RSAPublicKey) uniquePEMTypeMethod() {}

// DSAParameters is the enum type for 'valueDSAParameters' value
type DSAParameters struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueDSAParameters'
func (DSAParameters) New() PEMTypeEnum { return DSAParameters{pEMTypeEnumBase{valueDSAParameters}} }

// String returns always "DSAParameters" for this enum type
func (DSAParameters) String() string { return "DSAParameters" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (DSAParameters) uniquePEMTypeMethod() {}

// CMS is the enum type for 'valueCMS' value
type CMS struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueCMS'
func (CMS) New() PEMTypeEnum { return CMS{pEMTypeEnumBase{valueCMS}} }

// String returns always "CMS" for this enum type
func (CMS) String() string { return "CMS" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (CMS) uniquePEMTypeMethod() {}

// DSAPublicKey is the enum type for 'valueDSAPublicKey' value
type DSAPublicKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueDSAPublicKey'
func (DSAPublicKey) New() PEMTypeEnum { return DSAPublicKey{pEMTypeEnumBase{valueDSAPublicKey}} }

// String returns always "DSAPublicKey" for this enum type
func (DSAPublicKey) String() string { return "DSAPublicKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (DSAPublicKey) uniquePEMTypeMethod() {}

// X942DHParameters is the enum type for 'valueX942DHParameters' value
type X942DHParameters struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueX942DHParameters'
func (X942DHParameters) New() PEMTypeEnum { return X942DHParameters{pEMTypeEnumBase{valueX942DHParameters}} }

// String returns always "X942DHParameters" for this enum type
func (X942DHParameters) String() string { return "X942DHParameters" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (X942DHParameters) uniquePEMTypeMethod() {}

// SSLSessionParameters is the enum type for 'valueSSLSessionParameters' value
type SSLSessionParameters struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueSSLSessionParameters'
func (SSLSessionParameters) New() PEMTypeEnum { return SSLSessionParameters{pEMTypeEnumBase{valueSSLSessionParameters}} }

// String returns always "SSLSessionParameters" for this enum type
func (SSLSessionParameters) String() string { return "SSLSessionParameters" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (SSLSessionParameters) uniquePEMTypeMethod() {}

// PKCS6 is the enum type for 'valuePKCS6' value
type PKCS6 struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valuePKCS6'
func (PKCS6) New() PEMTypeEnum { return PKCS6{pEMTypeEnumBase{valuePKCS6}} }

// String returns always "PKCS6" for this enum type
func (PKCS6) String() string { return "PKCS6" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (PKCS6) uniquePEMTypeMethod() {}

// TrustedCertificate is the enum type for 'valueTrustedCertificate' value
type TrustedCertificate struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueTrustedCertificate'
func (TrustedCertificate) New() PEMTypeEnum { return TrustedCertificate{pEMTypeEnumBase{valueTrustedCertificate}} }

// String returns always "TrustedCertificate" for this enum type
func (TrustedCertificate) String() string { return "TrustedCertificate" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (TrustedCertificate) uniquePEMTypeMethod() {}

// DSAPrivateKey is the enum type for 'valueDSAPrivateKey' value
type DSAPrivateKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueDSAPrivateKey'
func (DSAPrivateKey) New() PEMTypeEnum { return DSAPrivateKey{pEMTypeEnumBase{valueDSAPrivateKey}} }

// String returns always "DSAPrivateKey" for this enum type
func (DSAPrivateKey) String() string { return "DSAPrivateKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (DSAPrivateKey) uniquePEMTypeMethod() {}

// DHParameters is the enum type for 'valueDHParameters' value
type DHParameters struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueDHParameters'
func (DHParameters) New() PEMTypeEnum { return DHParameters{pEMTypeEnumBase{valueDHParameters}} }

// String returns always "DHParameters" for this enum type
func (DHParameters) String() string { return "DHParameters" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (DHParameters) uniquePEMTypeMethod() {}

// OpenSSHPrivateKey is the enum type for 'valueOpenSSHPrivateKey' value
type OpenSSHPrivateKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueOpenSSHPrivateKey'
func (OpenSSHPrivateKey) New() PEMTypeEnum { return OpenSSHPrivateKey{pEMTypeEnumBase{valueOpenSSHPrivateKey}} }

// String returns always "OpenSSHPrivateKey" for this enum type
func (OpenSSHPrivateKey) String() string { return "OpenSSHPrivateKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (OpenSSHPrivateKey) uniquePEMTypeMethod() {}

// Parameters is the enum type for 'valueParameters' value
type Parameters struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueParameters'
func (Parameters) New() PEMTypeEnum { return Parameters{pEMTypeEnumBase{valueParameters}} }

// String returns always "Parameters" for this enum type
func (Parameters) String() string { return "Parameters" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Parameters) uniquePEMTypeMethod() {}

// NewCertificateRequest is the enum type for 'valueNewCertificateRequest' value
type NewCertificateRequest struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueNewCertificateRequest'
func (NewCertificateRequest) New() PEMTypeEnum { return NewCertificateRequest{pEMTypeEnumBase{valueNewCertificateRequest}} }

// String returns always "NewCertificateRequest" for this enum type
func (NewCertificateRequest) String() string { return "NewCertificateRequest" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (NewCertificateRequest) uniquePEMTypeMethod() {}

// PrivateKey is the enum type for 'valuePrivateKey' value
type PrivateKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valuePrivateKey'
func (PrivateKey) New() PEMTypeEnum { return PrivateKey{pEMTypeEnumBase{valuePrivateKey}} }

// String returns always "PrivateKey" for this enum type
func (PrivateKey) String() string { return "PrivateKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (PrivateKey) uniquePEMTypeMethod() {}

// ECPrivateKey is the enum type for 'valueECPrivateKey' value
type ECPrivateKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueECPrivateKey'
func (ECPrivateKey) New() PEMTypeEnum { return ECPrivateKey{pEMTypeEnumBase{valueECPrivateKey}} }

// String returns always "ECPrivateKey" for this enum type
func (ECPrivateKey) String() string { return "ECPrivateKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (ECPrivateKey) uniquePEMTypeMethod() {}

// ECDSAPublicKey is the enum type for 'valueECDSAPublicKey' value
type ECDSAPublicKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueECDSAPublicKey'
func (ECDSAPublicKey) New() PEMTypeEnum { return ECDSAPublicKey{pEMTypeEnumBase{valueECDSAPublicKey}} }

// String returns always "ECDSAPublicKey" for this enum type
func (ECDSAPublicKey) String() string { return "ECDSAPublicKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (ECDSAPublicKey) uniquePEMTypeMethod() {}

// X509Certificate is the enum type for 'valueX509Certificate' value
type X509Certificate struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueX509Certificate'
func (X509Certificate) New() PEMTypeEnum { return X509Certificate{pEMTypeEnumBase{valueX509Certificate}} }

// String returns always "X509Certificate" for this enum type
func (X509Certificate) String() string { return "X509Certificate" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (X509Certificate) uniquePEMTypeMethod() {}

// RSAPrivateKey is the enum type for 'valueRSAPrivateKey' value
type RSAPrivateKey struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valueRSAPrivateKey'
func (RSAPrivateKey) New() PEMTypeEnum { return RSAPrivateKey{pEMTypeEnumBase{valueRSAPrivateKey}} }

// String returns always "RSAPrivateKey" for this enum type
func (RSAPrivateKey) String() string { return "RSAPrivateKey" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (RSAPrivateKey) uniquePEMTypeMethod() {}

// PKCS6SignedData is the enum type for 'valuePKCS6SignedData' value
type PKCS6SignedData struct{ pEMTypeEnumBase }

// New is the constructor for a brand new PEMTypeEnum with value 'valuePKCS6SignedData'
func (PKCS6SignedData) New() PEMTypeEnum { return PKCS6SignedData{pEMTypeEnumBase{valuePKCS6SignedData}} }

// String returns always "PKCS6SignedData" for this enum type
func (PKCS6SignedData) String() string { return "PKCS6SignedData" }

// uniquePEMTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (PKCS6SignedData) uniquePEMTypeMethod() {}

var internalPEMTypeEnumValues = []PEMTypeEnum{
	X509CRL{}.New(),
	PublicKey{}.New(),
	EncryptedPrivateKey{}.New(),
	ECParameters{}.New(),
	Certificate{}.New(),
	CertificateRequest{}.New(),
	PGPPrivateKeyBlock{}.New(),
	AnyPrivateKey{}.New(),
	RSAPublicKey{}.New(),
	DSAParameters{}.New(),
	CMS{}.New(),
	DSAPublicKey{}.New(),
	X942DHParameters{}.New(),
	SSLSessionParameters{}.New(),
	PKCS6{}.New(),
	TrustedCertificate{}.New(),
	DSAPrivateKey{}.New(),
	DHParameters{}.New(),
	OpenSSHPrivateKey{}.New(),
	Parameters{}.New(),
	NewCertificateRequest{}.New(),
	PrivateKey{}.New(),
	ECPrivateKey{}.New(),
	ECDSAPublicKey{}.New(),
	X509Certificate{}.New(),
	RSAPrivateKey{}.New(),
	PKCS6SignedData{}.New(),
}

// PEMTypeEnumValues will return a slice of all allowed enum value types
func PEMTypeEnumValues() []PEMTypeEnum { return internalPEMTypeEnumValues[:] }

// NewPEMTypeFromValue will generate a valid enum from a value, or return nil in case of invalid value
func NewPEMTypeFromValue(v string) (result PEMTypeEnum) {
	switch v {
	case valueX509CRL:
		result = X509CRL{}.New()
	case valuePublicKey:
		result = PublicKey{}.New()
	case valueEncryptedPrivateKey:
		result = EncryptedPrivateKey{}.New()
	case valueECParameters:
		result = ECParameters{}.New()
	case valueCertificate:
		result = Certificate{}.New()
	case valueCertificateRequest:
		result = CertificateRequest{}.New()
	case valuePGPPrivateKeyBlock:
		result = PGPPrivateKeyBlock{}.New()
	case valueAnyPrivateKey:
		result = AnyPrivateKey{}.New()
	case valueRSAPublicKey:
		result = RSAPublicKey{}.New()
	case valueDSAParameters:
		result = DSAParameters{}.New()
	case valueCMS:
		result = CMS{}.New()
	case valueDSAPublicKey:
		result = DSAPublicKey{}.New()
	case valueX942DHParameters:
		result = X942DHParameters{}.New()
	case valueSSLSessionParameters:
		result = SSLSessionParameters{}.New()
	case valuePKCS6:
		result = PKCS6{}.New()
	case valueTrustedCertificate:
		result = TrustedCertificate{}.New()
	case valueDSAPrivateKey:
		result = DSAPrivateKey{}.New()
	case valueDHParameters:
		result = DHParameters{}.New()
	case valueOpenSSHPrivateKey:
		result = OpenSSHPrivateKey{}.New()
	case valueParameters:
		result = Parameters{}.New()
	case valueNewCertificateRequest:
		result = NewCertificateRequest{}.New()
	case valuePrivateKey:
		result = PrivateKey{}.New()
	case valueECPrivateKey:
		result = ECPrivateKey{}.New()
	case valueECDSAPublicKey:
		result = ECDSAPublicKey{}.New()
	case valueX509Certificate:
		result = X509Certificate{}.New()
	case valueRSAPrivateKey:
		result = RSAPrivateKey{}.New()
	case valuePKCS6SignedData:
		result = PKCS6SignedData{}.New()
	}
	return
}

// MustGetPEMTypeFromValue is the same as NewPEMTypeFromValue, but will panic in case of conversion failure
func MustGetPEMTypeFromValue(v string) PEMTypeEnum {
	result := NewPEMTypeFromValue(v)
	if result == nil {
		panic("invalid PEMTypeEnum value cast")
	}
	return result
}
