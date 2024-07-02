{
  inputs = {
    nix-vscode-extensions.url = "github:nix-community/nix-vscode-extensions";
    flake-utils.follows = "nix-vscode-extensions/flake-utils";
    nixpkgs.follows = "nix-vscode-extensions/nixpkgs";
  };

  outputs = { nixpkgs, flake-utils, nix-vscode-extensions, ... }: flake-utils.lib.eachDefaultSystem (
    system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
      extensions = nix-vscode-extensions.extensions.${system};
      inherit (pkgs) vscode-with-extensions vscodium;
    in
    rec {
      packages.default = vscode-with-extensions.override {
        vscode = vscodium;
        vscodeExtensions = with extensions.vscode-marketplace; [
          eamodio.gitlens
          github.copilot
          github.copilot-chat
          golang.go
          jnoortheen.nix-ide
        ];
      };

      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          packages.default
          go
        ];
      };
    }
  );
}
