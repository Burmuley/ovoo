package services

// import (
// 	"context"
// 	"errors"
// 	"reflect"
// 	"testing"

// 	"github.com/Burmuley/ovoo/internal/entities"
// 	"github.com/Burmuley/ovoo/internal/repositories"
// )

// func TestUsersUsecase_GetById(t *testing.T) {
// 	type fields struct {
// 		repo repositories.UsersReadWriter
// 	}
// 	type args struct {
// 		id string
// 	}
// 	fld := fields{repo: createTestUserRepo()}
// 	tests := []struct {
// 		name          string
// 		fields        fields
// 		args          args
// 		want          *entities.User
// 		wantErr       bool
// 		wantErrTarget error
// 	}{
// 		{
// 			name:          "Non-existing id",
// 			fields:        fld,
// 			args:          args{id: entities.NewId()},
// 			want:          nil,
// 			wantErr:       true,
// 			wantErrTarget: ErrNotFound,
// 		},
// 		{
// 			name:          "Invalid id",
// 			fields:        fld,
// 			args:          args{id: "invalidID"},
// 			want:          nil,
// 			wantErr:       true,
// 			wantErrTarget: ErrValidation,
// 		},
// 		{
// 			name:          "Existing id",
// 			fields:        fld,
// 			args:          args{id: preCreateOwners[0].ID},
// 			want:          &preCreateOwners[0],
// 			wantErr:       false,
// 			wantErrTarget: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			u := &UsersUsecase{
// 				repo: tt.fields.repo,
// 			}
// 			got, err := u.GetById(context.Background(), tt.args.id)
// 			if (err != nil) != tt.wantErr || !errors.Is(err, tt.wantErrTarget) {
// 				t.Errorf("UsersUsecase.GetById() error = %v, wantErr %v, wantErrTarget %v", err, tt.wantErr, tt.wantErrTarget)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("UsersUsecase.GetById() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestUsersUsecase_GetByLogin(t *testing.T) {
// 	type fields struct {
// 		repo repositories.UsersReadWriter
// 	}
// 	type args struct {
// 		login string
// 	}
// 	fld := fields{repo: createTestUserRepo()}
// 	tests := []struct {
// 		name          string
// 		fields        fields
// 		args          args
// 		want          *entities.User
// 		wantErr       bool
// 		wantErrTarget error
// 	}{
// 		{
// 			name:          "Non-existing login",
// 			fields:        fld,
// 			args:          args{login: "some-email@non-exitent-domain.com"},
// 			want:          nil,
// 			wantErr:       true,
// 			wantErrTarget: ErrNotFound,
// 		},
// 		{
// 			name:          "Invalid login",
// 			fields:        fld,
// 			args:          args{login: "invalidLogin"},
// 			want:          nil,
// 			wantErr:       true,
// 			wantErrTarget: ErrValidation,
// 		},
// 		{
// 			name:          "Existing login",
// 			fields:        fld,
// 			args:          args{login: preCreateOwners[0].Login},
// 			want:          &preCreateOwners[0],
// 			wantErr:       false,
// 			wantErrTarget: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			u := &UsersUsecase{
// 				repo: tt.fields.repo,
// 			}
// 			got, err := u.GetByLogin(context.Background(), tt.args.login)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("UsersUsecase.GetByLogin() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("UsersUsecase.GetByLogin() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestUsersUsecase_Create(t *testing.T) {
// 	type fields struct {
// 		repo repositories.UsersReadWriter
// 	}
// 	type args struct {
// 		user *entities.User
// 	}
// 	fld := fields{repo: createTestUserRepo()}
// 	tests := []struct {
// 		name          string
// 		fields        fields
// 		args          args
// 		want          *entities.User
// 		wantErr       bool
// 		wantErrTarget error
// 	}{
// 		{
// 			name:   "Invalid login",
// 			fields: fld,
// 			args: args{
// 				user: &entities.User{
// 					FirstName: "Test F Name",
// 					LastName:  "Test L Name",
// 					Login:     "invalidLogin",
// 				},
// 			},
// 			want:          nil,
// 			wantErr:       true,
// 			wantErrTarget: ErrValidation,
// 		},
// 		{
// 			name:   "Valid login",
// 			fields: fld,
// 			args: args{
// 				user: &entities.User{
// 					FirstName: "Test F Name",
// 					LastName:  "Test L Name",
// 					Login:     "some-test@login.com",
// 				},
// 			},
// 			want: &entities.User{
// 				FirstName: "Test F Name",
// 				LastName:  "Test L Name",
// 				Login:     "some-test@login.com",
// 			},
// 			wantErr:       false,
// 			wantErrTarget: nil,
// 		},
// 		// TODO: add validation tests with empty fields
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			u := &UsersUsecase{
// 				repo: tt.fields.repo,
// 			}
// 			got, err := u.Create(context.Background(), tt.args.user)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("UsersUsecase.Create() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got == nil && tt.wantErr == true {
// 				return
// 			}
// 			tt.want.ID = got.ID
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("UsersUsecase.Create() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestUsersUsecase_Delete(t *testing.T) {
// 	type fields struct {
// 		repo repositories.UsersReadWriter
// 	}
// 	type args struct {
// 		id string
// 	}
// 	fld := fields{repo: createTestUserRepo()}
// 	tests := []struct {
// 		name          string
// 		fields        fields
// 		args          args
// 		wantErr       bool
// 		wantErrTarget error
// 	}{
// 		{
// 			name:          "Non-exitent ID",
// 			fields:        fld,
// 			args:          args{id: entities.NewId()},
// 			wantErr:       true,
// 			wantErrTarget: ErrNotFound,
// 		},
// 		{
// 			name:          "Invalid ID",
// 			fields:        fld,
// 			args:          args{id: "invalidId"},
// 			wantErr:       true,
// 			wantErrTarget: ErrValidation,
// 		},
// 		{
// 			name:   "Existing ID",
// 			fields: fld,
// 			args: func() args {
// 				user := &entities.User{
// 					ID:        entities.NewId(),
// 					FirstName: "Test User FN",
// 					LastName:  "Test User LN",
// 					Login:     "test-user-to-delete@test.local",
// 				}
// 				nu, _ := fld.repo.Create(context.Background(), user)
// 				return args{id: nu.ID}
// 			}(),
// 			wantErr:       false,
// 			wantErrTarget: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			u := &UsersUsecase{
// 				repo: tt.fields.repo,
// 			}
// 			if err := u.Delete(context.Background(), tt.args.id); (err != nil) != tt.wantErr {
// 				t.Errorf("UsersUsecase.Delete() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
