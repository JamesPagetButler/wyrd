import Lake
open Lake DSL

package wyrd where
  -- add package configuration options here

require mathlib from git
  "https://github.com/leanprover-community/mathlib4.git" @ "a090f46da78e9af11fee348cd7ee47bf8dd219d2"

@[default_target]
lean_lib Wyrd where
  -- add library configuration options here
