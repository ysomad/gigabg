package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_scaleCopies(t *testing.T) {
	t.Parallel()

	type args struct {
		base       int
		players int
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "8 players", args: args{base: 15, players: 8}, want: 15},
		{name: "4 players", args: args{base: 15, players: 4}, want: 7},
		{name: "2 players", args: args{base: 15, players: 2}, want: 3},
		{name: "2 players low base", args: args{base: 5, players: 2}, want: 1},
		{name: "min 1", args: args{base: 1, players: 1}, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scaleCopies(tt.args.base, tt.args.players))
		})
	}
}
