{
    description = "lazymake";

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
        flake-utils.url = "github:numtide/flake-utils";
    };

    outputs = { self, nixpkgs, flake-utils }:
        flake-utils.lib.eachDefaultSystem (system:
                let
                pkgs = import nixpkgs {
                inherit system;
                };
                in {
                packages.default = pkgs.buildGoModule {
                pname = "lazymake";
                version = "0.4.1";

                src = ./.;

                vendorHash = "sha256-X/n7eoughxIP42JcLfifnbyqjYzRQBGsQvvCvFElotY=";

                ldflags = [
                "-s"
                "-w"
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
                });
}
