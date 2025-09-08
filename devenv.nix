{ pkgs, lib, config, inputs, ... }:

{
  # --- Environnement ---
  # Variables d'env chargées automatiquement par direnv/devenv
  env = {
    # HTTP
    PORT = "8080";

    # Postgres (devenv)
    POSTGRES_USER = "postgres";
    POSTGRES_PASSWORD = "postgres";
    POSTGRES_DB = "projectdb";
    # Utiliser 5433 pour éviter les collisions avec Docker/Podman (5432)
    DB_PORT = "5433";

    # DSN de l'app quand elle tourne en dehors des conteneurs
    DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:5433/projectdb?sslmode=disable";

  # Assure que go utilise /tmp pour les binaires temporaires (évite noexec sur /run/user)
  TMPDIR = "/tmp";

    # Démo
    GREET = "devenv";
  };

  # https://devenv.sh/packages/
  packages = [
    pkgs.git
    pkgs.postgresql
    pkgs.go
  ];

  # https://devenv.sh/languages/
  # languages.rust.enable = true;

  # https://devenv.sh/processes/
  # processes.cargo-watch.exec = "cargo-watch";

  # --- Services ---
  # Postgres local géré par devenv
  services.postgres = {
    enable = true;
    package = pkgs.postgresql_15;
    listen_addresses = "127.0.0.1";
    port = lib.toInt config.env.DB_PORT; # 5433

    initialDatabases = [
        {
      name = config.env.POSTGRES_DB;
        }
      ]; # projectdb
  };

  # https://devenv.sh/scripts/
  scripts.hello.exec = ''
    echo hello from $GREET
  '';

  enterShell = ''
    hello
    git --version
  echo "Postgres on 127.0.0.1:$DB_PORT (db=$POSTGRES_DB user=$POSTGRES_USER)"
  echo "DATABASE_URL=$DATABASE_URL"
  '';

  # https://devenv.sh/tasks/
  # tasks = {
  #   "myproj:setup".exec = "mytool build";
  #   "devenv:enterShell".after = [ "myproj:setup" ];
  # };

  # https://devenv.sh/tests/
  enterTest = ''
    echo "Running tests"
    git --version | grep --color=auto "${pkgs.git.version}"
  '';

  # https://devenv.sh/git-hooks/
  # git-hooks.hooks.shellcheck.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
