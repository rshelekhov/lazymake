{
  description = "Interactive TUI for browsing and executing Makefile targets";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = self.shortRev or self.dirtyShortRev or "dev";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "lazymake";
          inherit version;

          src = ./.;

          # Update when go.mod/go.sum change: set to pkgs.lib.fakeHash,
          # run `nix build`, and copy the hash from the error.
          vendorHash = "sha256-X/n7eoughxIP42JcLfifnbyqjYzRQBGsQvvCvFElotY=";

          subPackages = [ "cmd/lazymake" ];

          ldflags = [
            "-s"
            "-w"
            "-X github.com/rshelekhov/lazymake/version.Version=${version}"
          ];
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/lazymake";
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
          ];
        };
      }
    );
}
