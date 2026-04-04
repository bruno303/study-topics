package transaction

import "testing"

func TestJoin(t *testing.T) {
	opts := Join()

	if opts.Propagation != PropagationJoin {
		t.Fatalf("expected propagation %v, got %v", PropagationJoin, opts.Propagation)
	}
}

func TestRequiresNew(t *testing.T) {
	opts := RequiresNew()

	if opts.Propagation != PropagationRequiresNew {
		t.Fatalf("expected propagation %v, got %v", PropagationRequiresNew, opts.Propagation)
	}
}

func TestNested(t *testing.T) {
	opts := Nested()

	if opts.Propagation != PropagationNested {
		t.Fatalf("expected propagation %v, got %v", PropagationNested, opts.Propagation)
	}
}

func TestEmptyOpts_DefaultsToRequiresNew(t *testing.T) {
	if EmptyOpts.Propagation != PropagationRequiresNew {
		t.Fatalf("expected empty opts propagation %v, got %v", PropagationRequiresNew, EmptyOpts.Propagation)
	}
}

func TestOpts_EffectivePropagation(t *testing.T) {
	tests := []struct {
		name string
		opts Opts
		want Propagation
	}{
		{
			name: "explicit join",
			opts: Opts{Propagation: PropagationJoin},
			want: PropagationJoin,
		},
		{
			name: "explicit requires new",
			opts: Opts{Propagation: PropagationRequiresNew},
			want: PropagationRequiresNew,
		},
		{
			name: "explicit nested",
			opts: Opts{Propagation: PropagationNested},
			want: PropagationNested,
		},
		{
			name: "unspecified propagation defaults to join",
			opts: Opts{},
			want: PropagationJoin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.EffectivePropagation()
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
