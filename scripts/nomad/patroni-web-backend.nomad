job "patroni-web-backend" {
  datacenters = [
    "nomad-eu",
  ]

  type = "service"

  update {
    max_parallel     = 1
    health_check     = "checks"
    min_healthy_time = "30s"
    healthy_deadline = "5m"
    auto_revert      = true
    canary           = 1
    stagger          = "15s"
  }

  migrate {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "30s"
    healthy_deadline = "5m"
  }

  reschedule {
    attempts       = 15
    interval       = "1h"
    delay          = "30s"
    delay_function = "exponential"
    max_delay      = "120s"
    unlimited      = false
  }

  group "backend" {
    count = 1

    restart {
      attempts = 5
      delay = "30s"
    }

    task "backend" {
      driver = "docker"

      config {
        image        = "<DOCKER_IMAGE_NAME>:latest"
        network_mode = "macvlan_net"
        command      = "/opt/patroni-web-backend/bin/patroni-web-backend"
      }

      vault {
        policies = [ "read" ]
      }

      env {
        WEB_UI_URL = "<PATRONI_WEB_UI_URL>",
        PB_PORT = "80",
        CONSUL_HTTP_ADDR = "<CONSUL_HTTP_ADDR>:8500",
        CONSUL_DC = "infra1",
        CONSUL_KV = "services/patroni-ui/config",
        PGSQL_DB = "patron",
        PGSQL_USER = "patron",
        PGSQL_PORT = "5432",
        PGSQL_HOST = "<POSTGRES_HOST>",
        PGSQL_PASS = "<PG_PASS>"
        ELASTIC_HOST = "<ELASTIC_HOST>:9200",
        ELASTIC_LOG = "true"
        IPA_AUTH = "true"
        IPA_HOST = "<FREE_IPA_HOST>"
      }

      service {
        address_mode = "driver"
        name         = "patroni-web-back"
        port         = "80"

        check {
          address_mode = "driver"
          port         = "80"
          type         = "tcp"
          interval     = "10s"
          timeout      = "5s"
        }

        check_restart {
          limit           = 5
          grace           = "90s"
          ignore_warnings = false
        }
      }

      resources {
        cpu    = 100
        memory = 512

        network {
          mbits = 10
        }
      }
    }
  }
}
