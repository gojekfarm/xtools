package xload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_collisionSyncMap_err(t *testing.T) {
	tests := []struct {
		name    string
		cm      func() *collisionSyncMap
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty keys",
			cm: func() *collisionSyncMap {
				m := &collisionSyncMap{}
				m.add("")
				m.add("")
				m.add("")
				return m
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.cm().err(), "err()")
		})
	}
}
