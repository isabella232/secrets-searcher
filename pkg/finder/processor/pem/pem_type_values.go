package pem

//go:generate sh -c "go run github.com/gdm85/go-genums PEMType value string pem_type_values.go > pem_type.go"

const (
    valueX509Certificate       = "X509 CERTIFICATE"
    valueCertificate           = "CERTIFICATE"
    valueTrustedCertificate    = "TRUSTED CERTIFICATE"
    valueNewCertificateRequest = "NEW CERTIFICATE REQUEST"
    valueCertificateRequest    = "CERTIFICATE REQUEST"
    valueX509CRL               = "X509 CRL"
    valueAnyPrivateKey         = "ANY PRIVATE KEY"
    valuePublicKey             = "PUBLIC KEY"
    valueRSAPrivateKey         = "RSA PRIVATE KEY"
    valueRSAPublicKey          = "RSA PUBLIC KEY"
    valueDSAPrivateKey         = "DSA PRIVATE KEY"
    valueDSAPublicKey          = "DSA PUBLIC KEY"
    valuePKCS6                 = "PKCS7"
    valuePKCS6SignedData       = "PKCS #7 SIGNED DATA"
    valueEncryptedPrivateKey   = "ENCRYPTED PRIVATE KEY"
    valuePrivateKey            = "PRIVATE KEY"
    valueDHParameters          = "DH PARAMETERS"
    valueX942DHParameters      = "X9.42 DH PARAMETERS"
    valueSSLSessionParameters  = "SSL SESSION PARAMETERS"
    valueDSAParameters         = "DSA PARAMETERS"
    valueECDSAPublicKey        = "ECDSA PUBLIC KEY"
    valueECParameters          = "EC PARAMETERS"
    valueECPrivateKey          = "EC PRIVATE KEY"
    valueParameters            = "PARAMETERS"
    valueCMS                   = "CMS"

    // Not sure why these aren't here:
    // https://github.com/openssl/openssl/blob/1531241/include/openssl/pem.h#L32-L56
    valueOpenSSHPrivateKey  = "OPENSSH PRIVATE KEY"
    valuePGPPrivateKeyBlock = "PGP PRIVATE KEY BLOCK"
)
