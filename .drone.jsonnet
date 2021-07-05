local golang_image = 'golang:1.16-alpine';

local common = {
  kind: 'pipeline',
  type: 'docker',
  platform: {
    os: 'linux',
    arch: 'amd64',
  },
};

local volume_mount = { name: 'dockersock', host: { path: '/var/run/docker.sock' } };
local volume_step_mount = { name: 'dockersock', path: '/var/run/docker.sock' };

local lint_pipeline = common {
  name: 'lint',
  steps: [
    {
      name: 'golangci-lint',
      image: 'golangci/golangci-lint',
      commands: [
        'golangci-lint run -E gosec ./...',
      ],
    },
  ],
};

local build_pipeline(arch) = common {
  name: 'build %s' % arch,
  steps: [
    {
      name: 'build',
      image: golang_image,
      environment: {
        GOOS: 'linux',
        GOARCH: arch,
      },
      commands: [
        'go build ./cmd/cloudflare-ddns/main.go',
      ],
    },
  ],
  depends_on: ['lint'],
};

local docker_pipeline(branch, tag) = common {
  name: 'docker image %s' % branch,
  trigger: {
    branch: (if branch == '*' then { exclude: ['main', 'develop'] } else branch),
  },
  steps: [
    {
      name: 'build and publish',
      image: 'thegeeklab/drone-docker-buildx',
      environment: {
        DOCKER_TOKEN: {
          from_secret: 'DOCKER_TOKEN',
        },
      },
      volumes: [volume_step_mount],
      commands: [
        'echo "$DOCKER_TOKEN" | docker login ghcr.io -u gordonpn --password-stdin',
        'docker run --rm --privileged multiarch/qemu-user-static --reset -p yes',
        'docker buildx rm builder || true',
        'docker buildx create --name builder --driver docker-container --use',
        'docker buildx inspect --bootstrap',
        'docker buildx build -t ghcr.io/gordonpn/cloudflare-ddns:%s --platform linux/amd64,linux/arm64,linux/arm/v7 --push .' % tag,
      ],
    },
  ],
  volumes: [volume_mount],
  depends_on: ['build amd64', 'build arm64', 'build arm'],
};

local deploy_prod = common {
  name: 'deploy prod',
  trigger: {
    branch: 'main',
    event: 'push',
  },
  steps: [
    {
      name: 'deploy prod',
      image: 'docker/compose:1.29.2',
      environment: {
        API_TOKEN: {
          from_secret: 'API_TOKEN',
        },
        ZONE_ID: {
          from_secret: 'ZONE_ID',
        },
        HC_URL: {
          from_secret: 'HC_URL',
        },
      },
      volumes: [volume_step_mount],
      commands: [
        'docker-compose -f /drone/src/docker-compose.yml config > /drone/src/docker-compose.processed.yml',
        'docker stack deploy -c /drone/src/docker-compose.processed.yml cloudflare-ddns',
      ],
    },
  ],
  volumes: [volume_mount],
  depends_on: ['docker image main'],
};

[
  lint_pipeline,
  build_pipeline('amd64'),
  build_pipeline('arm64'),
  build_pipeline('arm'),
  docker_pipeline('main', 'stable'),
  docker_pipeline('develop', 'latest'),
  docker_pipeline('*', '${DRONE_COMMIT_SHA}'),
  deploy_prod,
]
