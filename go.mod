module profile

go 1.14

require (
	github.com/dadrus/gin-authn v0.0.0-20200707075323-5aac0009266d
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.6.3
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
)

// replace github.com/dadrus/gin-authn => ../gin-authn
