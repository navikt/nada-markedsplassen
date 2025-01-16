package iamcredentials_test

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/navikt/nada-backend/pkg/iamcredentials"
	"github.com/navikt/nada-backend/pkg/iamcredentials/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"testing"
)

const publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAr0Jswk8lh4IOaWhpsa0i
JhZt0RS0NNyFrRLBZhkuuQP1nns+c3gnYssG4esWyKrJrma0McHBYRhtZZALB40Q
G2UXFVjNmStcotTh/iNpdDBAnoaYax7iq88IDBh5pqE3FrpPrjRdnDd/dIrx6lJS
zvaN5jZAw8x26GKN4HyiRvEfc9TNcbgrDuld84XcnEqZdMreJrkO6X9gxf4vpBM9
lPAuDqMFUn5DvQCvp9PI4Fal+vm3zDThFD/lE6bp/K/cau93C++XN47vhJn7oYV0
utg1eWFqGmxnsO3I85slL2F0BBIEC74bNyGu/8bXEPtP4fKiSO/2VEtzrEYG2/H/
twIDAQAB
-----END PUBLIC KEY-----
`

const privateKey = `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCvQmzCTyWHgg5p
aGmxrSImFm3RFLQ03IWtEsFmGS65A/Weez5zeCdiywbh6xbIqsmuZrQxwcFhGG1l
kAsHjRAbZRcVWM2ZK1yi1OH+I2l0MECehphrHuKrzwgMGHmmoTcWuk+uNF2cN390
ivHqUlLO9o3mNkDDzHboYo3gfKJG8R9z1M1xuCsO6V3zhdycSpl0yt4muQ7pf2DF
/i+kEz2U8C4OowVSfkO9AK+n08jgVqX6+bfMNOEUP+UTpun8r9xq73cL75c3ju+E
mfuhhXS62DV5YWoabGew7cjzmyUvYXQEEgQLvhs3Ia7/xtcQ+0/h8qJI7/ZUS3Os
Rgbb8f+3AgMBAAECggEADpwuTjHO4nGtxeJrEowTprKD9m66vxVgchci1uIOimSR
cI68RrVOLebYQgkZCHgZrAI/z04O/YsjGNkIh66eGHU1lr/6aQ8Q/+St75kAIiwg
7EDdf+s+i4K3cbAF+XrDCY//3c6GZzlxKe6opWy2HoQLPD/APRJUb0xCoOOC8QA7
7cQX0E8n9+n1J76TzBDQgJnBBYC2G5V4eGdYgVgDsy3kaWpkFYsLTQWeSSeO8+qZ
QKPw6Yu0jQs812ajUUcMl53tKsZkpYJAnsPodTTw3lq5oJFF6XJ9Jf89DzhIayro
e7A/83zaDGlVfj5LxNSSF37ykKbToh/3ykefMuEioQKBgQDvF1Mu7jCptQH+hqff
h3idNJERYDNBjQ8m14mK1+mf5gz4czCeBtVY0WkDsobk8PHj+OtgawPxhun75AfA
qFwVKAQ1zoYP9rZDpqMkNys8Wiig7OBUEj4FbnGmnFP+kTEeG9LYxPHLA5ssPtGj
ycrrt76bY75Jr0cCEAvOqHWHTwKBgQC7p3JU/ZhsbwZHsqqnZSMd1BjcZfoPIrqy
TK4Eguo/GW4dfgOHsWnWtLHJ0TokZ6bdyJVoFciYLImFuUiA1SNNXlp1aBrNk24k
TnX00jeytax9DzIHsOHCS6ahrdbteiPLRhWTgC1MrLKPxM9erZq3IaBcAgZFS0mk
JIR/fslnGQKBgDLZMRXADpVpK51oIffGJf65GUkqvnvodhp6qIPg24zoLkYAqYxS
Q7l5/+2LYGj8XVVwsQ52dAY//S9XFdcBd2QAeLTA0X4/qA/HNtcS7J0PR6jB+Aup
PYuGK6GVib+QPXP70uHLMOlOQQgt7AP7fK6ZC26czfF51442v2waI7S9AoGAMpqI
GWU9mlgiQGls3bFHU/7jKWQSl8xMvlIxRyQqmRN5f1iBCTGNkgmuO/dBD5ooBHzX
1XayXl78QuRhKeTQHUgJasnFGJTeScoiwv+BZ57YQe08F5jaeHPAHq9rWyTpzCI9
JUaWcKvNhzmSljyIkUPvI4CkQkF4PVxfoqYFF9kCgYBUmL3Vx0rvE7h6ZKKlKJhL
LOUOLmXXOUsc64A5WeIpW7kuMaJJG3xeTQUjTe8py0Kif/UuXgawNSYNYLKKhNtO
HinoLzcgmwn4soz/TKYJXZOk1Czorxb1WH3uELTCh49fpwYIstJxuurB8sLqfMQn
TlNiEcMXKC9tdGiqi6SstA==
-----END PRIVATE KEY-----
`

func TestSignJWT(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx     context.Context
		signer  *iamcredentials.ServiceAccount
		payload jwt.MapClaims
	}
	tests := []struct {
		name    string
		args    args
		want    *iamcredentials.SignedJWT
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				signer: &iamcredentials.ServiceAccount{Email: "test@test-project.iam.gserviceaccount.com"},
				payload: jwt.MapClaims{
					"ident": "test",
					"ip":    "10.10.10.10",
				},
			},
			want: &iamcredentials.SignedJWT{
				SignedJWT: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZGVudCI6InRlc3QiLCJpcCI6IjEwLjEwLjEwLjEwIn0.hF0-1n6NeaIsq6T7ZkYk9u3qKVXpS8n5vSQybrXPwast3f9KexVagwa4ltjPUD2zFlQpsFyyrYcReqZq1Q9P1U8lw8NYSMqw0SEvCFoZ03SMDNH5LuOHyQ_xmu2mFj0v4F2P-xTKzNSFfFAEPVZm3zqbrnyfotTRALMDXYTIKSbUEEXX3uZYu_Zu99zR5sohetubHrb9fYghyYWRsDuxyKkROWMd_FWnoryDemrFtnjL_HF2oujuNnjqrwbDvhAU2FNxD6sLCTLPmv8IFJJZkcdH1PYQ5f-GEj9aOSeZ9rwlf85ehLEkABSpkCmbH6_zdZzRHx1iwQ8J7u0YpPQBew",
				KeyID:     "ee9c39b5d6182ba8afa1dba7388ec2108d2b2d11",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			log := zerolog.New(zerolog.NewConsoleWriter())
			e := emulator.New(log)
			e.AddSigner("test@test-project.iam.gserviceaccount.com", publicKey, privateKey)
			url := e.Run()

			c := iamcredentials.New(url, true)

			got, err := c.SignJWT(tc.args.ctx, tc.args.signer, tc.args.payload)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)

				token, err := jwt.Parse(got.SignedJWT, func(token *jwt.Token) (interface{}, error) {
					pub, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
					if err != nil {
						return nil, err
					}

					return pub, nil
				})

				require.NoError(t, err)
				claims, ok := token.Claims.(jwt.MapClaims)
				require.True(t, ok)
				require.Equal(t, tc.args.payload, claims)

			}
		})
	}
}
