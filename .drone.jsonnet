local golang_image = 'golang:1.16-alpine';

local common = {
  kind: 'pipeline',
  type: 'docker',
  platform: {
    os: 'linux',
    arch: 'amd64',
  },
};

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
        // DOCKER_BUILDKIT: 1,
      },
      volumes: [
        {
          name: 'dockersock',
          path: '/var/run/docker.sock',
        },
      ],
      commands: [
        'echo "$DOCKER_TOKEN" | docker login ghcr.io -u gordonpn --password-stdin',
        'docker run --rm --privileged multiarch/qemu-user-static --reset -p yes',
        'docker buildx rm builder || true',
        'docker buildx create --name builder --driver docker-container --use',
        'docker buildx inspect --bootstrap',
        'docker buildx build -t ghcr.io/gordonpn/cloudflare-ddns:%s --platform linux/amd64,linux/arm64 --push .' % tag,
      ],
    },
  ],
  volumes: [
    {
      name: 'dockersock',
      host: {
        path: '/var/run/docker.sock',
      },
    },
  ],
};

[
  lint_pipeline,
  build_pipeline('amd64'),
  build_pipeline('arm64'),
  build_pipeline('arm'),
  docker_pipeline('main', 'stable'),
  docker_pipeline('develop', 'latest'),
  docker_pipeline('*', '${DRONE_COMMIT_SHA}'),
]
