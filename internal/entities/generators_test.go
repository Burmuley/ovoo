package entities

import (
	"regexp"
	"testing"
)

var testDict = []string{"soooome", "words", "toooooo", "teeeeest", "theeeeee", "function", "oopsie"}

func TestGenAliasEmail(t *testing.T) {
	type args struct {
		domain    string
		wordsDict []string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid email",
			args: args{
				wordsDict: testDict,
				domain:    "aliases-test.local",
			},
			wantErr: false,
			want:    `\w+-\w+-\w{3}@aliases-test.local`,
		},
		{
			name: "Empty domain parameter",
			args: args{
				wordsDict: testDict,
				domain:    "",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "Nil dictionary parameter",
			args: args{
				wordsDict: nil,
				domain:    "aliases-test.local",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "Empty dictionary parameter",
			args: args{
				wordsDict: []string{},
				domain:    "aliases-test.local",
			},
			wantErr: true,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenAliasEmail(tt.args.domain, tt.args.wordsDict)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenAliasEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			rx := regexp.MustCompile(tt.want)
			if !rx.MatchString(got.String()) {
				t.Errorf("GenReplyAliasEmail() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateReplyAliasEmail(t *testing.T) {
	type args struct {
		senderEmail Email
		aliasEmail  Email
		domain      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple email string",
			args: args{
				senderEmail: "sender@domain.com",
				aliasEmail:  "alias1@protected.local",
				domain:      "protected.local",
			},
			wantErr: false,
			want:    `sender_at_domain_com_\w{10}_\w{8}@protected.local`,
		},
		{
			name: "complex email string #1",
			args: args{
				senderEmail: "some_other@email.from.complex.domain.org",
				aliasEmail:  "alias1@protected.local",
				domain:      "protected.local",
			},
			wantErr: false,
			want:    `some_other_at_email_from_complex_domain_org_\w{10}_\w{8}@protected.local`,
		},
		{
			name: "complex email string #2",
			args: args{
				senderEmail: "some.more@email.domain-withdash.org",
				aliasEmail:  "alias1@protected.local",
				domain:      "protected.local",
			},
			wantErr: false,
			want:    `some_more_at_email_domain_withdash_org_\w{10}_\w{8}@protected.local`,
		},
		{
			name: "invalid sender email #1",
			args: args{
				senderEmail: "some.more@email@domain-withdash.org",
				aliasEmail:  "alias1@protected.local",
				domain:      "protected.local",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "invalid sender email #2",
			args: args{
				senderEmail: "some.more#email.domain-withdash.org",
				aliasEmail:  "alias1@protected.local",
				domain:      "protected.local",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "invalid alias email #1",
			args: args{
				senderEmail: "sender@domain.com",
				aliasEmail:  "alias1@protected@local",
				domain:      "protected.local",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "invalid alias email #2",
			args: args{
				senderEmail: "sender@domain.com",
				aliasEmail:  "alias1#protected.local",
				domain:      "protected.local",
			},
			wantErr: true,
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repeat := 2
			gots := make([]Email, 0, repeat)
			hashes := make([]Hash, 0, repeat)

			for range repeat {
				got, hash, err := GenReplyAliasEmail(tt.args.senderEmail, tt.args.aliasEmail, tt.args.domain)
				if (err != nil) != tt.wantErr {
					t.Errorf("GenReplyAliasEmail() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				rx := regexp.MustCompile(tt.want)
				if !rx.MatchString(got.String()) {
					t.Errorf("GenReplyAliasEmail() got = %v, want %v", got, tt.want)
				}
				if err := hash.Validate(); err != nil && tt.wantErr == false {
					t.Errorf("GenReplyAliasEmail() hash validation: %s", err.Error())
				}
				gots = append(gots, got)
				hashes = append(hashes, hash)
			}

			for i := 1; i < repeat; i++ {
				if gots[i-1].String() != gots[i].String() {
					t.Errorf("GenReplyAliasEmail() reply alias email values different between iterations: %v", gots)
				}

				if hashes[i-1].String() != hashes[i].String() {
					t.Errorf("GenReplyAliasEmail() reply alias email values different between iterations: %v", gots)
				}
			}
		})
	}
}
