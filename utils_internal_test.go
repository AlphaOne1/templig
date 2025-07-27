package templig

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err        error
		text       string
		wantErr    bool
		wantErrMsg string
	}{
		{ // 0
			text:       "wrap this",
			err:        errors.New("original error"),
			wantErr:    true,
			wantErrMsg: "wrap this: original error",
		},
		{ // 1
			text:       "wrap that: %w",
			err:        errors.New("original error"),
			wantErr:    true,
			wantErrMsg: "wrap that: %w: original error",
		},
		{ // 2
			text:    "wrap this",
			err:     nil,
			wantErr: false,
		},
		{ // 3
			text:       "",
			err:        errors.New("error case"),
			wantErr:    true,
			wantErrMsg: "error case",
		},
		{ // 4
			text:    "",
			err:     nil,
			wantErr: false,
		},
	}

	for testIndex, v := range tests {
		t.Run(fmt.Sprintf("WrapError-%d", testIndex), func(t *testing.T) {
			t.Parallel()

			got := wrapError(v.text, v.err)

			if (got != nil) != v.wantErr {
				t.Errorf(`%v: got error "%v", but wanted "%v"`, testIndex, got, v.wantErr)
			}

			if v.wantErr && got.Error() != v.wantErrMsg {
				t.Errorf(`%v: got error "%v" but wanted "%v"`, testIndex, got.Error(), v.wantErrMsg)
			}
		})
	}
}
