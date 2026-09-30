package auth

import (
	"github.com/google/uuid"
	"net/http"
	"testing"
)

func TestMakeJWT(t *testing.T) {
	id := uuid.New()
	tokenSecret := "secret"
	jwtToken, err := MakeJWT(id, tokenSecret)
	if err != nil {
		t.Fatalf("Error in making JWT Token: %v", err)
	}

	jwtId, err := ValidateJWT(jwtToken, tokenSecret)
	if err != nil {
		t.Fatalf("Error in validating JWT Token: %v", err)
	}

	if jwtId != id {
		t.Fatalf("jwtId is not equal to id")
	}

}

func TestBearerToken(t *testing.T) {
	id := uuid.New()
	tokenSecret := "secret"
	_, err := MakeJWT(id, tokenSecret)
	if err != nil {
		t.Fatalf("Error in making JWT Token: %v", err)
	}

	newHttp := http.Request{
		Header: http.Header{"Authorization": []string{"Bearer test"}},
	}

	_, err = GetBearerToken(newHttp.Header)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}
