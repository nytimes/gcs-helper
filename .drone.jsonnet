// the first version is used to build the binary that gets shipped to Docker Hub.
local go_versions = ['1.12.8', '1.13beta1'];

local test_ci_dockerfile = {
  name: 'test-ci-dockerfile',
  image: 'plugins/docker',
  settings: {
    repo: 'nytimes/gcs-helper',
    dry_run: true,
  },
  when: {
    event: ['pull_request'],
  },
  depends_on: ['build'],
};

local push_to_dockerhub = {
  name: 'build-and-push-to-dockerhub',
  image: 'plugins/docker',
  settings: {
    repo: 'nytimes/gcs-helper',
    auto_tag: true,
    username: { from_secret: 'docker_username' },
    password: { from_secret: 'docker_password' },
  },
  when: {
    ref: [
      'refs/tags/*',
      'refs/heads/master',
    ],
  },
  depends_on: ['test', 'lint', 'build'],
};

local docker_sanity_check(name, image_version, refs) = {
  name: 'docker-sanity-check-%(name)s' % { name: name },
  image: 'nytimes/gcs-helper:%(image_version)s' % { image_version: image_version },
  pull: 'always',
  commands: ['gcs-helper -version'],
  when: {
    ref: refs,
  },
  depends_on: ['build-and-push-to-dockerhub'],
};

local dockerfile_steps = [
  test_ci_dockerfile,
  push_to_dockerhub,
  docker_sanity_check('push', 'latest', ['refs/heads/master']),
  docker_sanity_check('tag', '${DRONE_TAG#v}', ['refs/tags/*']),
];

local mod_download(go_version) = {
  name: 'mod-download',
  image: 'golang:%(go_version)s' % { go_version: go_version },
  commands: ['go mod download'],
  environment: { GOPROXY: 'https://proxy.golang.org' },
  depends_on: ['clone'],
};

local tests(go_version) = {
  name: 'test',
  image: 'golang:%(go_version)s' % { go_version: go_version },
  commands: ['go test -race -vet all -mod readonly ./...'],
  depends_on: ['mod-download'],
};

local lint = {
  name: 'lint',
  image: 'golangci/golangci-lint',
  pull: 'always',
  commands: ['golangci-lint run --enable-all -D errcheck -D lll -D dupl -D gochecknoglobals --deadline 5m ./...'],
  depends_on: ['mod-download'],
};

local build(go_version) = {
  name: 'build',
  image: 'golang:%(go_version)s' % { go_version: go_version },
  commands: ['go build -o gcs-helper -mod readonly -ldflags "-X main.version=$${DRONE_TAG:-SNAPSHOT-${DRONE_COMMIT}}"'],
  environment: { CGO_ENABLED: 0 },
  depends_on: ['mod-download'],
};

local sanity_check = {
  name: 'sanity-check',
  image: 'alpine',
  commands: ['./gcs-helper -version'],
  depends_on: ['build'],
};

local pipeline(go_version) = {
  kind: 'pipeline',
  name: 'build_go_%(go_version)s' % { go_version: go_version },
  workspace: {
    base: '/go',
    path: 'gcs-helper-%(go_version)s' % { go_version: go_version },
  },
  steps: [
    mod_download(go_version),
    tests(go_version),
    lint,
    build(go_version),
    sanity_check,
  ] + if go_version == go_versions[0] then dockerfile_steps else [],
};

std.map(pipeline, go_versions)
