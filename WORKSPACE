workspace(name = "kubevirt")

# register crosscompiler toolchains
load("//bazel/toolchain:toolchain.bzl", "register_all_toolchains")

register_all_toolchains()

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load(
    "@bazel_tools//tools/build_defs/repo:http.bzl",
    "http_archive",
    "http_file",
)

http_archive(
    name = "rules_python",
    sha256 = "934c9ceb552e84577b0faf1e5a2f0450314985b4d8712b2b70717dc679fdc01b",
    urls = [
        "https://github.com/bazelbuild/rules_python/releases/download/0.3.0/rules_python-0.3.0.tar.gz",
        "https://storage.googleapis.com/builddeps/934c9ceb552e84577b0faf1e5a2f0450314985b4d8712b2b70717dc679fdc01b",
    ],
)

http_archive(
    name = "rules_oci",
    sha256 = "acbf8f40e062f707f8754e914dcb0013803c6e5e3679d3e05b571a9f5c7e0b43",
    strip_prefix = "rules_oci-2.0.1",
    urls = [
        "https://github.com/bazel-contrib/rules_oci/releases/download/v2.0.1/rules_oci-v2.0.1.tar.gz",
        "https://storage.googleapis.com/builddeps/acbf8f40e062f707f8754e914dcb0013803c6e5e3679d3e05b571a9f5c7e0b43",
    ],
)

load("@rules_oci//oci:dependencies.bzl", "rules_oci_dependencies")

rules_oci_dependencies()

load("@rules_oci//oci:repositories.bzl", "oci_register_toolchains")

oci_register_toolchains(
    name = "oci",
)

load("@rules_oci//oci:pull.bzl", "oci_pull")

# Bazel buildtools prebuilt binaries
http_archive(
    name = "buildifier_prebuilt",
    sha256 = "7f85b688a4b558e2d9099340cfb510ba7179f829454fba842370bccffb67d6cc",
    strip_prefix = "buildifier-prebuilt-7.3.1",
    urls = [
        "http://github.com/keith/buildifier-prebuilt/archive/7.3.1.tar.gz",
        "https://storage.googleapis.com/builddeps/7f85b688a4b558e2d9099340cfb510ba7179f829454fba842370bccffb67d6cc",
    ],
)

load("@buildifier_prebuilt//:deps.bzl", "buildifier_prebuilt_deps")

buildifier_prebuilt_deps()

# Additional bazel rules

http_archive(
    name = "platforms",
    sha256 = "3384eb1c30762704fbe38e440204e114154086c8fc8a8c2e3e28441028c019a8",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/platforms/releases/download/1.0.0/platforms-1.0.0.tar.gz",
        "https://github.com/bazelbuild/platforms/releases/download/1.0.0/platforms-1.0.0.tar.gz",
        "https://storage.googleapis.com/builddeps/3384eb1c30762704fbe38e440204e114154086c8fc8a8c2e3e28441028c019a8",
    ],
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "130739704540caa14e77c54810b9f01d6d9ae897d53eedceb40fd6b75efc3c23",
    urls = [
        "https://github.com/bazel-contrib/rules_go/releases/download/v0.54.1/rules_go-v0.54.1.zip",
        "https://storage.googleapis.com/builddeps/130739704540caa14e77c54810b9f01d6d9ae897d53eedceb40fd6b75efc3c23",
    ],
)

load("@buildifier_prebuilt//:defs.bzl", "buildifier_prebuilt_register_toolchains", "buildtools_assets")

buildifier_prebuilt_register_toolchains(
    assets = buildtools_assets(
        arches = [
            "amd64",
            "arm64",
            "s390x",
        ],
        names = [
            "buildifier",
            "buildozer",
        ],
        platforms = [
            "darwin",
            "linux",
            "windows",
        ],
        sha256_values = {
            "buildifier_darwin_amd64": "375f823103d01620aaec20a0c29c6cbca99f4fd0725ae30b93655c6704f44d71",
            "buildifier_darwin_arm64": "5a6afc6ac7a09f5455ba0b89bd99d5ae23b4174dc5dc9d6c0ed5ce8caac3f813",
            "buildifier_linux_amd64": "5474cc5128a74e806783d54081f581662c4be8ae65022f557e9281ed5dc88009",
            "buildifier_linux_arm64": "0bf86c4bfffaf4f08eed77bde5b2082e4ae5039a11e2e8b03984c173c34a561c",
            "buildifier_linux_s390x": "e2d79ff5885d45274f76531f1adbc7b73a129f59e767f777e8fbde633d9d4e2e",
            "buildifier_windows_amd64": "370cd576075ad29930a82f5de132f1a1de4084c784a82514bd4da80c85acf4a8",
            "buildozer_darwin_amd64": "854c9583efc166602276802658cef3f224d60898cfaa60630b33d328db3b0de2",
            "buildozer_darwin_arm64": "31b1bfe20d7d5444be217af78f94c5c43799cdf847c6ce69794b7bf3319c5364",
            "buildozer_linux_amd64": "3305e287b3fcc68b9a35fd8515ee617452cd4e018f9e6886b6c7cdbcba8710d4",
            "buildozer_linux_arm64": "0b5a2a717ac4fc911e1fec8d92af71dbb4fe95b10e5213da0cc3d56cea64a328",
            "buildozer_linux_s390x": "7e28da8722656e800424989f5cdbc095cb29b2d398d33e6b3d04e0f50bc0bb10",
            "buildozer_windows_amd64": "58d41ce53257c5594c9bc86d769f580909269f68de114297f46284fbb9023dcf",
        },
        version = "v7.3.1",
    ),
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "b760f7fe75173886007f7c2e616a21241208f3d90e8657dc65d36a771e916b6a",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.39.1/bazel-gazelle-v0.39.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.39.1/bazel-gazelle-v0.39.1.tar.gz",
        "https://storage.googleapis.com/builddeps/b760f7fe75173886007f7c2e616a21241208f3d90e8657dc65d36a771e916b6a",
    ],
)

http_archive(
    name = "rules_pkg",
    sha256 = "d20c951960ed77cb7b341c2a59488534e494d5ad1d30c4818c736d57772a9fef",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/1.0.1/rules_pkg-1.0.1.tar.gz",
        "https://github.com/bazelbuild/rules_pkg/releases/download/1.0.1/rules_pkg-1.0.1.tar.gz",
        "https://storage.googleapis.com/builddeps/d20c951960ed77cb7b341c2a59488534e494d5ad1d30c4818c736d57772a9fef",
    ],
)

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "95d39fd84ff4474babaf190450ee034d958202043e366b9fc38f438c9e6c3334",
    strip_prefix = "rules_docker-0.16.0",
    urls = [
        "https://github.com/bazelbuild/rules_docker/releases/download/v0.16.0/rules_docker-v0.16.0.tar.gz",
        "https://storage.googleapis.com/builddeps/95d39fd84ff4474babaf190450ee034d958202043e366b9fc38f438c9e6c3334",
    ],
)

http_archive(
    name = "com_github_ash2k_bazel_tools",
    sha256 = "46fdbc00930c8dc9d84690b5bd94db6b4683b061199967d2cda1cfbda8f02c49",
    strip_prefix = "bazel-tools-19b174803c0db1a01e77f10fa2079c35f54eed6e",
    urls = [
        "https://github.com/ash2k/bazel-tools/archive/19b174803c0db1a01e77f10fa2079c35f54eed6e.zip",
        "https://storage.googleapis.com/builddeps/46fdbc00930c8dc9d84690b5bd94db6b4683b061199967d2cda1cfbda8f02c49",
    ],
)

# Disk images
http_file(
    name = "alpine_image",
    sha256 = "f87a0fd3ab0e65d2a84acd5dad5f8b6afce51cb465f65dd6f8a3810a3723b6e4",
    urls = [
        "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-virt-3.20.1-x86_64.iso",
        "https://storage.googleapis.com/builddeps/f87a0fd3ab0e65d2a84acd5dad5f8b6afce51cb465f65dd6f8a3810a3723b6e4",
    ],
)

http_file(
    name = "alpine_image_aarch64",
    sha256 = "ca2f0e8aa7a1d7917bce7b9e7bd413772b64ec529a1938d20352558f90a5035a",
    urls = [
        "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/aarch64/alpine-virt-3.20.1-aarch64.iso",
        "https://storage.googleapis.com/builddeps/ca2f0e8aa7a1d7917bce7b9e7bd413772b64ec529a1938d20352558f90a5035a",
    ],
)

http_file(
    name = "alpine_image_s390x",
    sha256 = "4ca1462252246d53e4949523b87fcea088e8b4992dbd6df792818c5875069b16",
    urls = [
        "https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/s390x/alpine-standard-3.18.8-s390x.iso",
        "https://storage.googleapis.com/builddeps/4ca1462252246d53e4949523b87fcea088e8b4992dbd6df792818c5875069b16",
    ],
)

http_file(
    name = "cirros_image",
    sha256 = "932fcae93574e242dc3d772d5235061747dfe537668443a1f0567d893614b464",
    urls = [
        "https://download.cirros-cloud.net/0.5.2/cirros-0.5.2-x86_64-disk.img",
        "https://storage.googleapis.com/builddeps/932fcae93574e242dc3d772d5235061747dfe537668443a1f0567d893614b464",
    ],
)

http_file(
    name = "cirros_image_aarch64",
    sha256 = "889c1117647b3b16cfc47957931c6573bf8e755fc9098fdcad13727b6c9f2629",
    urls = [
        "https://download.cirros-cloud.net/0.5.2/cirros-0.5.2-aarch64-disk.img",
        "https://storage.googleapis.com/builddeps/889c1117647b3b16cfc47957931c6573bf8e755fc9098fdcad13727b6c9f2629",
    ],
)

http_file(
    name = "virtio_win_image",
    sha256 = "57b0f6dc8dc92dc2ae8621f8b1bfbd8a873de9bedc788c4c4b305ea28acc77cd",
    urls = [
        "https://fedorapeople.org/groups/virt/virtio-win/direct-downloads/archive-virtio/virtio-win-0.1.266-1/virtio-win-0.1.266.iso",
        "https://storage.googleapis.com/builddeps/57b0f6dc8dc92dc2ae8621f8b1bfbd8a873de9bedc788c4c4b305ea28acc77cd",
    ],
)

http_archive(
    name = "bazeldnf",
    sha256 = "0a4b9740da1839ded674c8f5012c069b235b101f1eaa2552a4721287808541af",
    strip_prefix = "bazeldnf-v0.5.9-2",
    urls = [
        "https://github.com/brianmcarey/bazeldnf/releases/download/v0.5.9-2/bazeldnf-v0.5.9-2.tar.gz",
        "https://storage.googleapis.com/builddeps/0a4b9740da1839ded674c8f5012c069b235b101f1eaa2552a4721287808541af",
    ],
)

load("@bazeldnf//bazeldnf:defs.bzl", "rpm")
load(
    "@bazeldnf//bazeldnf:repositories.bzl",
    "bazeldnf_dependencies",
    "bazeldnf_register_toolchains",
)
load(
    "@io_bazel_rules_go//go:deps.bzl",
    "go_register_toolchains",
    "go_rules_dependencies",
)

bazeldnf_dependencies()

bazeldnf_register_toolchains(
    name = "bazeldnf_prebuilt",
)

go_rules_dependencies()

go_register_toolchains(
    go_version = "1.24.9",
    nogo = "@//:nogo_vet",
)

load("@com_github_ash2k_bazel_tools//goimports:deps.bzl", "goimports_dependencies")

goimports_dependencies()

load(
    "@bazel_gazelle//:deps.bzl",
    "gazelle_dependencies",
    "go_repository",
)

gazelle_dependencies(go_sdk = "go_sdk")

bazeldnf_dependencies()

# Winrmcli dependencies
go_repository(
    name = "com_github_masterzen_winrmcli",
    commit = "c85a68ee8b6e3ac95af2a5fd62d2f41c9e9c5f32",
    importpath = "github.com/masterzen/winrm-cli",
)

# Winrmcp deps
go_repository(
    name = "com_github_packer_community_winrmcp",
    commit = "c76d91c1e7db27b0868c5d09e292bb540616c9a2",
    importpath = "github.com/packer-community/winrmcp",
)

go_repository(
    name = "com_github_masterzen_winrm_cli",
    commit = "6f0c57dee4569c04f64c44c335752b415e5d73a7",
    importpath = "github.com/masterzen/winrm-cli",
)

go_repository(
    name = "com_github_masterzen_winrm",
    commit = "1d17eaf15943ca3554cdebb3b1b10aaa543a0b7e",
    importpath = "github.com/masterzen/winrm",
)

go_repository(
    name = "com_github_nu7hatch_gouuid",
    commit = "179d4d0c4d8d407a32af483c2354df1d2c91e6c3",
    importpath = "github.com/nu7hatch/gouuid",
)

go_repository(
    name = "com_github_dylanmei_iso8601",
    commit = "2075bf119b58e5576c6ed9f867b8f3d17f2e54d4",
    importpath = "github.com/dylanmei/iso8601",
)

go_repository(
    name = "com_github_gofrs_uuid",
    commit = "abfe1881e60ef34074c1b8d8c63b42565c356ed6",
    importpath = "github.com/gofrs/uuid",
)

go_repository(
    name = "com_github_christrenkamp_goxpath",
    commit = "c5096ec8773dd9f554971472081ddfbb0782334e",
    importpath = "github.com/ChrisTrenkamp/goxpath",
)

go_repository(
    name = "com_github_azure_go_ntlmssp",
    commit = "4a21cbd618b459155f8b8ee7f4491cd54f5efa77",
    importpath = "github.com/Azure/go-ntlmssp",
)

go_repository(
    name = "com_github_masterzen_simplexml",
    commit = "31eea30827864c9ab643aa5a0d5b2d4988ec8409",
    importpath = "github.com/masterzen/simplexml",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "4def268fd1a49955bfb3dda92fe3db4f924f2285",
    importpath = "golang.org/x/crypto",
)

# override rules_docker issue with this dependency
# rules_docker 0.16 uses 0.1.4, let's grab by commit
go_repository(
    name = "com_github_google_go_containerregistry",
    commit = "8a2841911ffee4f6892ca0083e89752fb46c48dd",  # v0.1.4
    importpath = "github.com/google/go-containerregistry",
)

# Pull go_image_base
oci_pull(
    name = "go_image_base",
    digest = "sha256:d5f7dca58e3db53d1de502bd1a747ecb1110cf6b0773af129f951ee11e2e3ed4",
    image = "gcr.io/distroless/base-debian12",
)

oci_pull(
    name = "go_image_base_aarch64",
    digest = "sha256:ba2aeab48a1dadbd47ac4ce37b7f6084043a8f59172f8b73a3ede3c3e1a71be4",
    image = "gcr.io/distroless/base-debian12",
)

oci_pull(
    name = "go_image_base_s390x",
    digest = "sha256:214b82df32d6dfe855715b7ce56dfe72a777da2c1e0b9fe47efb8cbc5cce5484",
    image = "gcr.io/distroless/base-debian12",
)

# Pull fedora container-disk preconfigured with ci tooling
# like stress and qemu guest agent pre-configured
# TODO build fedora_with_test_tooling for multi-arch
oci_pull(
    name = "fedora_with_test_tooling",
    digest = "sha256:897af945d1c58366086d5933ae4f341a5f1413b88e6c7f2b659436adc5d0f522",
    image = "quay.io/kubevirtci/fedora-with-test-tooling",
)

oci_pull(
    name = "alpine_with_test_tooling",
    digest = "sha256:8c8e8bb6cd81c75e492c678abb3e5f186d52eba2174ebabc328316250acfea58",
    image = "quay.io/kubevirtci/alpine-with-test-tooling-container-disk",
)

oci_pull(
    name = "alpine_with_test_tooling_s390x",
    digest = "sha256:1a52903133c00507607e8a82308a34923e89288d852762b9f4d5da227767e965",
    image = "quay.io/kubevirtci/alpine-with-test-tooling-container-disk",
)

oci_pull(
    name = "fedora_with_test_tooling_aarch64",
    digest = "sha256:3d5a2a95f7f9382dc6730073fe19a6b1bc668b424c362339c88c6a13dff2ef49",
    image = "quay.io/kubevirtci/fedora-with-test-tooling",
)

oci_pull(
    name = "fedora_with_test_tooling_s390x",
    digest = "sha256:3d9f468750d90845a81608ea13c85237ea295c6295c911a99dc5e0504c8bc05b",
    image = "quay.io/kubevirtci/fedora-with-test-tooling",
)

oci_pull(
    name = "alpine-ext-kernel-boot-demo-container-base",
    digest = "sha256:de4bc8de772ff7570e6dda871ea9cdd502feeeff1973f16f84bfbd60ff8f4149",
    image = "quay.io/kubevirt/alpine-ext-kernel-boot-demo",
)

# TODO build fedora_realtime for multi-arch
oci_pull(
    name = "fedora_realtime",
    digest = "sha256:f91379d202a5493aba9ce06870b5d1ada2c112f314530c9820a9ad07426aa565",
    image = "quay.io/kubevirt/fedora-realtime-container-disk",
)

oci_pull(
    name = "busybox",
    digest = "sha256:545e6a6310a27636260920bc07b994a299b6708a1b26910cfefd335fdfb60d2b",
    image = "registry.k8s.io/busybox",
)

http_archive(
    name = "io_bazel_rules_container_rpm",
    sha256 = "151261f1b81649de6e36f027c945722bff31176f1340682679cade2839e4b1e1",
    strip_prefix = "rules_container_rpm-0.0.5",
    urls = [
        "https://github.com/rmohr/rules_container_rpm/archive/v0.0.5.tar.gz",
        "https://storage.googleapis.com/builddeps/151261f1b81649de6e36f027c945722bff31176f1340682679cade2839e4b1e1",
    ],
)

http_archive(
    name = "libguestfs-appliance-x86_64",
    sha256 = "fb6da700eeae24da89aae6516091f7c5f46958b0b7812d2b122dc11dca1ab26a",
    urls = [
        "https://storage.googleapis.com/kubevirt-prow/devel/release/kubevirt/libguestfs-appliance/libguestfs-appliance-1.54.0-qcow2-linux-5.14.0-575-centos9-amd64.tar.xz",
    ],
)

http_archive(
    name = "libguestfs-appliance-s390x",
    sha256 = "532cb951d4245265da645c8cce14033c19ea8f0d163c01e88f4153dae44e0f95",
    urls = [
        "https://storage.googleapis.com/kubevirt-prow/devel/release/kubevirt/libguestfs-appliance/libguestfs-appliance-1.54.0-qcow2-linux-5.14.0-575-centos9-s390x.tar.xz",
    ],
)

# Get container-disk-v1alpha RPM's
http_file(
    name = "qemu-img",
    sha256 = "669250ad47aad5939cf4d1b88036fd95a94845d8e0bbdb05e933f3d2fe262fea",
    urls = ["https://storage.googleapis.com/builddeps/669250ad47aad5939cf4d1b88036fd95a94845d8e0bbdb05e933f3d2fe262fea"],
)

# some repos which are not part of go_rules anymore
go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:oWX7TPOiFAMXLq8o0ikBYfCJVlRHBcsciT5bXOrH628=",
    version = "v0.0.0-20190311183353-d8887717615a",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    sum = "h1:g61tztE5qeGQ89tm6NTjjM9VPIm088od1l6aSorWRWg=",
    version = "v0.3.0",
)

register_toolchains("//:py_toolchain")

go_repository(
    name = "org_golang_x_mod",
    build_file_generation = "on",
    build_file_proto_mode = "disable",
    importpath = "golang.org/x/mod",
    sum = "h1:RM4zey1++hCTbCVQfnWeKs9/IEsaBLA8vTkd0WVtmH4=",
    version = "v0.3.0",
)

go_repository(
    name = "org_golang_x_xerrors",
    build_file_generation = "on",
    build_file_proto_mode = "disable",
    importpath = "golang.org/x/xerrors",
    sum = "h1:go1bK/D/BFZV2I8cIQd1NKEZ+0owSTG1fDTci4IqFcE=",
    version = "v0.0.0-20200804184101-5ec99f83aff1",
)

rpm(
    name = "acl-0__2.3.2-4.el10.aarch64",
    sha256 = "e5c1d6460330fabe5ef57fb4b13d46ab0840f93556d898b5179f1b267f34455f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/acl-2.3.2-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "acl-0__2.3.2-4.el10.s390x",
    sha256 = "295d62b3d46571e5327671616bff8d1872af066f41719e09d5e0554d00001e49",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/acl-2.3.2-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "acl-0__2.3.2-4.el10.x86_64",
    sha256 = "fd89f3c793d09fe633bf7721da719d29d599d01f65aaaa355b1b308a6fa580f2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/acl-2.3.2-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "alternatives-0__1.30-2.el10.aarch64",
    sha256 = "13d1cae28aecbc13bee2cf23391ec2ee41d39c51c9bb47f466fbad133d38f5c9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/alternatives-1.30-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "alternatives-0__1.30-2.el10.s390x",
    sha256 = "ab4f800759f602c25f483681b126b4eced6ba81331c9b613dd47a229379c71e1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/alternatives-1.30-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "alternatives-0__1.30-2.el10.x86_64",
    sha256 = "1c8b83bf3dd0fa8d998a3c801986f50ea3661c2f8a21c60971c0391c381919c8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/alternatives-1.30-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "audit-libs-0__4.0.3-5.el10.aarch64",
    sha256 = "f45973727e2dea77b2209bc9795c890abac187383a596b3cb81ab066b11ddb90",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/audit-libs-4.0.3-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "audit-libs-0__4.0.3-5.el10.s390x",
    sha256 = "1d7617a754258f58b0986c6f944621819381543eee344f60f348fc44bc2274c1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/audit-libs-4.0.3-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "audit-libs-0__4.0.3-5.el10.x86_64",
    sha256 = "a2be49cd9497b28aa9688b6e58bce216797c868559d249e3a08034e22d1e86f7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/audit-libs-4.0.3-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "augeas-libs-0__1.14.2-0.9.20260120gitf4135e3.el10.s390x",
    sha256 = "03bb6de04d7f5c64cf30fd0fe26301508b8fad5520c25fd1fc291132b01e0c7a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/augeas-libs-1.14.2-0.9.20260120gitf4135e3.el10.s390x.rpm",
    ],
)

rpm(
    name = "augeas-libs-0__1.14.2-0.9.20260120gitf4135e3.el10.x86_64",
    sha256 = "479f9bb17e1ede3ae449e0cc47474a935b519febdb65c214eedccc1e836aeb8b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/augeas-libs-1.14.2-0.9.20260120gitf4135e3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "authselect-0__1.5.0-8.el10.aarch64",
    sha256 = "6806edc3ab06e45d1077f5d89865ff94d6939004acf365f9eaaf407e02642666",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/authselect-1.5.0-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "authselect-0__1.5.0-8.el10.s390x",
    sha256 = "dad241106db112ab5cb7dcc45164af6a38c614739672d9f0136f4e85b5907d3e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/authselect-1.5.0-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "authselect-0__1.5.0-8.el10.x86_64",
    sha256 = "2a16d12c77181f77189fac10b4a3f76c2d0dd97e230d9074f7d24d2e4967ab35",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/authselect-1.5.0-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "authselect-libs-0__1.5.0-8.el10.aarch64",
    sha256 = "cc6557c5707792705ffe41b0deae2c76a30382a86e35d3cc812f7d872e9f5871",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/authselect-libs-1.5.0-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "authselect-libs-0__1.5.0-8.el10.s390x",
    sha256 = "890def3a93b6204476966957eee6a893adc9ef76f0212873fa240279575f4ad3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/authselect-libs-1.5.0-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "authselect-libs-0__1.5.0-8.el10.x86_64",
    sha256 = "c3cb5c662f1225e0c1f90c406c2ab3bfad8191a4d1b46614b49dfd298f33c53a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/authselect-libs-1.5.0-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "basesystem-0__11-22.el10.aarch64",
    sha256 = "76ff57f4d7565cd0e49f5e6dc38f3707dfe6a6b61317d883c2701be4277f2abf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/basesystem-11-22.el10.noarch.rpm",
    ],
)

rpm(
    name = "basesystem-0__11-22.el10.s390x",
    sha256 = "76ff57f4d7565cd0e49f5e6dc38f3707dfe6a6b61317d883c2701be4277f2abf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/basesystem-11-22.el10.noarch.rpm",
    ],
)

rpm(
    name = "basesystem-0__11-22.el10.x86_64",
    sha256 = "76ff57f4d7565cd0e49f5e6dc38f3707dfe6a6b61317d883c2701be4277f2abf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/basesystem-11-22.el10.noarch.rpm",
    ],
)

rpm(
    name = "bash-0__5.2.26-6.el10.aarch64",
    sha256 = "3f42c3de9fddc6e6c08f7c603ce29ed96d8d66f4425ce1c27bcb0d7d0e0490b5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/bash-5.2.26-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "bash-0__5.2.26-6.el10.s390x",
    sha256 = "07261872bd05c23366da7c2529b776dccfdf1a33c99d784370ebfde32d8909d7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/bash-5.2.26-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "bash-0__5.2.26-6.el10.x86_64",
    sha256 = "31eaf885847a6671a93e2b6e0d48e937ae5520f0442265aae19f4294260b5618",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/bash-5.2.26-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "binutils-0__2.41-60.el10.aarch64",
    sha256 = "829fe311199f54f58c0b0e5a8297b6b9c89ba7cce31e51b9a575ddd5f8aaf80f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/binutils-2.41-60.el10.aarch64.rpm",
    ],
)

rpm(
    name = "binutils-0__2.41-60.el10.s390x",
    sha256 = "cfb7608108550dc979945bf8bbbc99dc201933e81252020f662799c0fbe5c8a3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/binutils-2.41-60.el10.s390x.rpm",
    ],
)

rpm(
    name = "binutils-0__2.41-60.el10.x86_64",
    sha256 = "a948054c9e555dfd17d22864d345a42b948a6a60bdaea75f1324e48ec6aa285e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/binutils-2.41-60.el10.x86_64.rpm",
    ],
)

rpm(
    name = "binutils-gold-0__2.41-60.el10.aarch64",
    sha256 = "eb1c2c492ecbf6affd5d22d29b4b773fae418aa875b85af7c7187d5f975b58b4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/binutils-gold-2.41-60.el10.aarch64.rpm",
    ],
)

rpm(
    name = "binutils-gold-0__2.41-60.el10.s390x",
    sha256 = "0e9c9469fa781dfada84de48b56996aa185a72bcc207eb12cd4d97dff322fd4b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/binutils-gold-2.41-60.el10.s390x.rpm",
    ],
)

rpm(
    name = "binutils-gold-0__2.41-60.el10.x86_64",
    sha256 = "2ec08ee46033739b426b8275e6c0d271e833b471a8ef5658f57372963df1f44f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/binutils-gold-2.41-60.el10.x86_64.rpm",
    ],
)

rpm(
    name = "bzip2-0__1.0.8-25.el10.aarch64",
    sha256 = "30fd7d37e3f06d0b06b6f3e6fda58fd9d54582b0e497795719d81dc68ac88ba7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/bzip2-1.0.8-25.el10.aarch64.rpm",
    ],
)

rpm(
    name = "bzip2-0__1.0.8-25.el10.s390x",
    sha256 = "c9208b97a6a3e2cb7fc84a7bea4e330399cc6ef892c3f0abe20d5df10797eade",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/bzip2-1.0.8-25.el10.s390x.rpm",
    ],
)

rpm(
    name = "bzip2-0__1.0.8-25.el10.x86_64",
    sha256 = "ff7f8e9c3cc936d35033ec40545ee4a836db27c30c240d3aa39be4c8b0fda631",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/bzip2-1.0.8-25.el10.x86_64.rpm",
    ],
)

rpm(
    name = "bzip2-libs-0__1.0.8-25.el10.aarch64",
    sha256 = "ac836c2c133077d0e71092f2c21e69d3985ace8458af527440e13b7edf165beb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/bzip2-libs-1.0.8-25.el10.aarch64.rpm",
    ],
)

rpm(
    name = "bzip2-libs-0__1.0.8-25.el10.s390x",
    sha256 = "219adea56b92ecf22cb63fad38638e16115df270b78ea1fbd3cc1b183caf69a4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/bzip2-libs-1.0.8-25.el10.s390x.rpm",
    ],
)

rpm(
    name = "bzip2-libs-0__1.0.8-25.el10.x86_64",
    sha256 = "309c7dbb857254655c51c4ab02d8038137c1363058542d8701c9272609f5b433",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/bzip2-libs-1.0.8-25.el10.x86_64.rpm",
    ],
)

rpm(
    name = "ca-certificates-0__2025.2.80_v9.0.305-102.el10.aarch64",
    sha256 = "a5a8cf95b7cae489df2f6b4448b6d5100593256b0033376d25b2705985fad9dc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/ca-certificates-2025.2.80_v9.0.305-102.el10.noarch.rpm",
    ],
)

rpm(
    name = "ca-certificates-0__2025.2.80_v9.0.305-102.el10.s390x",
    sha256 = "a5a8cf95b7cae489df2f6b4448b6d5100593256b0033376d25b2705985fad9dc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/ca-certificates-2025.2.80_v9.0.305-102.el10.noarch.rpm",
    ],
)

rpm(
    name = "ca-certificates-0__2025.2.80_v9.0.305-102.el10.x86_64",
    sha256 = "a5a8cf95b7cae489df2f6b4448b6d5100593256b0033376d25b2705985fad9dc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/ca-certificates-2025.2.80_v9.0.305-102.el10.noarch.rpm",
    ],
)

rpm(
    name = "capstone-0__5.0.1-6.el10.aarch64",
    sha256 = "be12ff671fc1244c69b39284b61f4a7e825570d11176dcd83e8476010157db92",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/capstone-5.0.1-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "capstone-0__5.0.1-6.el10.s390x",
    sha256 = "f94850c0dedde1efd687de604a99f6461ec2cb394184f76e3d2d17af0654f0d0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/capstone-5.0.1-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "capstone-0__5.0.1-6.el10.x86_64",
    sha256 = "aa46343e831205d94b08f3d692f88b3a84a16f35b260152684ea10183d972160",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/capstone-5.0.1-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "centos-gpg-keys-0__10.0-19.el10.aarch64",
    sha256 = "9de24d7bd3ee5b686170e6f27bd99b6550d02a8d4df5d00a7c6a83750f4d4b0a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/centos-gpg-keys-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "centos-gpg-keys-0__10.0-19.el10.s390x",
    sha256 = "9de24d7bd3ee5b686170e6f27bd99b6550d02a8d4df5d00a7c6a83750f4d4b0a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/centos-gpg-keys-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "centos-gpg-keys-0__10.0-19.el10.x86_64",
    sha256 = "9de24d7bd3ee5b686170e6f27bd99b6550d02a8d4df5d00a7c6a83750f4d4b0a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/centos-gpg-keys-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "centos-stream-release-0__10.0-19.el10.aarch64",
    sha256 = "b47742c7d0ee92454c15b97bca9240b61de31547d7de039f67ba498703623188",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/centos-stream-release-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "centos-stream-release-0__10.0-19.el10.s390x",
    sha256 = "b47742c7d0ee92454c15b97bca9240b61de31547d7de039f67ba498703623188",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/centos-stream-release-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "centos-stream-release-0__10.0-19.el10.x86_64",
    sha256 = "b47742c7d0ee92454c15b97bca9240b61de31547d7de039f67ba498703623188",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/centos-stream-release-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "centos-stream-repos-0__10.0-19.el10.aarch64",
    sha256 = "5fa429468121be8530982d8776e69e7cf91f2c4f159d5152169898a451baf676",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/centos-stream-repos-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "centos-stream-repos-0__10.0-19.el10.s390x",
    sha256 = "5fa429468121be8530982d8776e69e7cf91f2c4f159d5152169898a451baf676",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/centos-stream-repos-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "centos-stream-repos-0__10.0-19.el10.x86_64",
    sha256 = "5fa429468121be8530982d8776e69e7cf91f2c4f159d5152169898a451baf676",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/centos-stream-repos-10.0-19.el10.noarch.rpm",
    ],
)

rpm(
    name = "coreutils-single-0__9.5-6.el10.aarch64",
    sha256 = "d1cfc460e243e2fc1934b8b0d173d2f2b37bb69b1eedef2cfdc93619cfe6998a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/coreutils-single-9.5-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "coreutils-single-0__9.5-6.el10.s390x",
    sha256 = "d6ba2511cf43ebd40110b9b1786923da409ff73ae90aac67a12121e9259beb49",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/coreutils-single-9.5-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "coreutils-single-0__9.5-6.el10.x86_64",
    sha256 = "b1f91efb9d930b8b021d3648610029f78433a65b39b578e69e575c9767be61d5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/coreutils-single-9.5-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "cpp-0__14.3.1-4.3.el10.aarch64",
    sha256 = "d88d1b7c37bc90ffaae0e729a8314cee7a2d3d3b6d24279fbf01c63c2c307408",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/cpp-14.3.1-4.3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "cpp-0__14.3.1-4.3.el10.s390x",
    sha256 = "d627790816fdaf878c633887d4f35ab3aeee8e703057db981f299246f286fba7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/cpp-14.3.1-4.3.el10.s390x.rpm",
    ],
)

rpm(
    name = "cpp-0__14.3.1-4.3.el10.x86_64",
    sha256 = "d173162b43fbf0948354cc90e68bbc37e31b026943f0ac1502a0367b535b7bb2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/cpp-14.3.1-4.3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "cracklib-0__2.9.11-8.el10.aarch64",
    sha256 = "04112224e2f1b7027ef15ee4cb9ede5bb89426b29f150692778d8f7ca155eea9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/cracklib-2.9.11-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "cracklib-0__2.9.11-8.el10.s390x",
    sha256 = "2e0c0ba830f1a497461b1a7f6e76f5d409c9bf87d2c4a6874957abe3fdb74be3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/cracklib-2.9.11-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "cracklib-0__2.9.11-8.el10.x86_64",
    sha256 = "4d648a415fe67550a22ff50befdaf9a33ccb55dbc9a2e3d4121ddfbe2ee843f7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/cracklib-2.9.11-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "cracklib-dicts-0__2.9.11-8.el10.aarch64",
    sha256 = "51210426186039c77239cbb3c710acbc9f7778ca44292204ffa2ecf1448e2c1e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/cracklib-dicts-2.9.11-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "cracklib-dicts-0__2.9.11-8.el10.s390x",
    sha256 = "45cf94fabce8c9c035df7db91b19fefec5cfef5cee54505cabebce1822e3099d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/cracklib-dicts-2.9.11-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "cracklib-dicts-0__2.9.11-8.el10.x86_64",
    sha256 = "79dd2684b0ae0cbc47739c0e292f17243eb448b92f74bae893cf1eb4aba14703",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/cracklib-dicts-2.9.11-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "crypto-policies-0__20251127-1.git27c2902.el10.aarch64",
    sha256 = "84f438e426f45ecf1ce51fc71a1bb4c1a1a1b5ee63faf793273ee5d6aaaecb33",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/crypto-policies-20251127-1.git27c2902.el10.noarch.rpm",
    ],
)

rpm(
    name = "crypto-policies-0__20251127-1.git27c2902.el10.s390x",
    sha256 = "84f438e426f45ecf1ce51fc71a1bb4c1a1a1b5ee63faf793273ee5d6aaaecb33",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/crypto-policies-20251127-1.git27c2902.el10.noarch.rpm",
    ],
)

rpm(
    name = "crypto-policies-0__20251127-1.git27c2902.el10.x86_64",
    sha256 = "84f438e426f45ecf1ce51fc71a1bb4c1a1a1b5ee63faf793273ee5d6aaaecb33",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/crypto-policies-20251127-1.git27c2902.el10.noarch.rpm",
    ],
)

rpm(
    name = "curl-0__8.12.1-4.el10.aarch64",
    sha256 = "7fe56b8ad3db9141cd721455717109785447e79358f4541d27bec012230db8c4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/curl-8.12.1-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "curl-0__8.12.1-4.el10.s390x",
    sha256 = "2aa147ae00c5fc1a0264f785127771e6ced0f4ee3d82a9bd6c48d1f240e44c7c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/curl-8.12.1-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "curl-0__8.12.1-4.el10.x86_64",
    sha256 = "30b38c7b64e1a33c6b69634fcb4b9d9f1714f9bd6530ee0175fc3be149f23d9b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/curl-8.12.1-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "cyrus-sasl-gssapi-0__2.1.28-27.el10.aarch64",
    sha256 = "f030977f59727e389143e1813c5fc848799abbea48ed60aca460dc2eb1a79637",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/cyrus-sasl-gssapi-2.1.28-27.el10.aarch64.rpm",
    ],
)

rpm(
    name = "cyrus-sasl-gssapi-0__2.1.28-27.el10.s390x",
    sha256 = "28c75a50cf3f092920ac56fb65805e9c875fc95d4e76bce0e1cc6b6d21e3fba3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/cyrus-sasl-gssapi-2.1.28-27.el10.s390x.rpm",
    ],
)

rpm(
    name = "cyrus-sasl-gssapi-0__2.1.28-27.el10.x86_64",
    sha256 = "f9ab02ca832fe4d5c1e1ee3abd7ff3db3815d164561350316032a82b44d68b6c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/cyrus-sasl-gssapi-2.1.28-27.el10.x86_64.rpm",
    ],
)

rpm(
    name = "cyrus-sasl-lib-0__2.1.28-27.el10.aarch64",
    sha256 = "917d6b8d2eff0dd71b55646c758b938ac7b9f0a298f2dffae5948c9865215067",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/cyrus-sasl-lib-2.1.28-27.el10.aarch64.rpm",
    ],
)

rpm(
    name = "cyrus-sasl-lib-0__2.1.28-27.el10.s390x",
    sha256 = "b40557a0d21461db27adf093fe6a72ec17a243f6743a3d1e26c32601753e97ee",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/cyrus-sasl-lib-2.1.28-27.el10.s390x.rpm",
    ],
)

rpm(
    name = "cyrus-sasl-lib-0__2.1.28-27.el10.x86_64",
    sha256 = "ea78a83980b03f3709266f5e4c96b41699fe8d5f7003fb9503c3a7529c6ca46a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/cyrus-sasl-lib-2.1.28-27.el10.x86_64.rpm",
    ],
)

rpm(
    name = "dbus-1__1.14.10-5.el10.aarch64",
    sha256 = "2f00025969ff8b32c254ec38919908120f83847e98285413c718d1ad0b2a8766",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/dbus-1.14.10-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "dbus-1__1.14.10-5.el10.s390x",
    sha256 = "2a746bab9a5c03b6bc2f680ad3be8ecf935404c17f6488de44e77ab61bdfedb8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/dbus-1.14.10-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "dbus-1__1.14.10-5.el10.x86_64",
    sha256 = "c71f38667ecebd3ba0adf415ccf181209330bb0e2ca9ad0bf4de9828b370b9e4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/dbus-1.14.10-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "dbus-broker-0__36-4.el10.aarch64",
    sha256 = "3716b1d4daa23c6fd965175473464ddfa91ea5651a68298a2e0b139021e23035",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/dbus-broker-36-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "dbus-broker-0__36-4.el10.s390x",
    sha256 = "3d1ec31218c8925602bb7fcd88150c628a0e24ab5cc4e7c63b85785202756283",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/dbus-broker-36-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "dbus-broker-0__36-4.el10.x86_64",
    sha256 = "a0778052571fe74351500a06e765219fcf53c0ca2eeb4969a2682a36ee9f9c10",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/dbus-broker-36-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "dbus-common-1__1.14.10-5.el10.aarch64",
    sha256 = "1cf5e00ed550daa874c5ec81be43f4606717a2465d72b733d3b9012015dfa751",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/dbus-common-1.14.10-5.el10.noarch.rpm",
    ],
)

rpm(
    name = "dbus-common-1__1.14.10-5.el10.s390x",
    sha256 = "1cf5e00ed550daa874c5ec81be43f4606717a2465d72b733d3b9012015dfa751",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/dbus-common-1.14.10-5.el10.noarch.rpm",
    ],
)

rpm(
    name = "dbus-common-1__1.14.10-5.el10.x86_64",
    sha256 = "1cf5e00ed550daa874c5ec81be43f4606717a2465d72b733d3b9012015dfa751",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/dbus-common-1.14.10-5.el10.noarch.rpm",
    ],
)

rpm(
    name = "dbus-libs-1__1.14.10-5.el10.aarch64",
    sha256 = "976a662683dc4f8235303cd6065f589c4d4728671116827b2002ac1fd4a74a72",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/dbus-libs-1.14.10-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "dbus-libs-1__1.14.10-5.el10.s390x",
    sha256 = "261a5aee8fd8417bdb0b629b7ae4141cec92de79d32b45982c66cc82878f3175",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/dbus-libs-1.14.10-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "dbus-libs-1__1.14.10-5.el10.x86_64",
    sha256 = "7cd5d99568a89ef7100ae60d44aa270cbf5882e95cbc8f43497696f81c664284",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/dbus-libs-1.14.10-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "device-mapper-10__1.02.206-3.el10.aarch64",
    sha256 = "185e448e20167139d421ce9177c63deaa75cd1a875110b4c0f2dd05481214141",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/device-mapper-1.02.206-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "device-mapper-10__1.02.206-3.el10.s390x",
    sha256 = "0da51b07ab865cae27da5d651513fb93358503e5160c1f034cf4db9e7393708e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/device-mapper-1.02.206-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "device-mapper-10__1.02.206-3.el10.x86_64",
    sha256 = "30fc452056e3b1117f3d48b88a7a9a1633690a9cb37b6c4c663ddd8b44dcf159",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/device-mapper-1.02.206-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "device-mapper-libs-10__1.02.206-3.el10.aarch64",
    sha256 = "8a6d569a6a478c816a4a596dde65e299e1072443e78237c5c44428a9557ed6bd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/device-mapper-libs-1.02.206-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "device-mapper-libs-10__1.02.206-3.el10.s390x",
    sha256 = "aa3c0b549b37445d7e5f754a09efd474ad385580ff4c4c612ca2eed3b2528fdd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/device-mapper-libs-1.02.206-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "device-mapper-libs-10__1.02.206-3.el10.x86_64",
    sha256 = "9342a7c107577149f860e71e4dd09002da1da5cf82e08d85a361758a6eaa6fee",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/device-mapper-libs-1.02.206-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "device-mapper-multipath-libs-0__0.9.9-15.el10.aarch64",
    sha256 = "14a7c7b2affa61a420527a77beba4b9968269b42671ef5bd0a690f11dd3241c8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/device-mapper-multipath-libs-0.9.9-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "device-mapper-multipath-libs-0__0.9.9-15.el10.x86_64",
    sha256 = "e88082ce08b8067cd35fd27daadddd11bd35d55be78f4e3985eb3074c52ef464",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/device-mapper-multipath-libs-0.9.9-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "diffutils-0__3.10-8.el10.aarch64",
    sha256 = "d06031d2cd612618343d29186bc873cafd52c9e71efae6d04dcb494de2b53b58",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/diffutils-3.10-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "diffutils-0__3.10-8.el10.s390x",
    sha256 = "4668ee01492723f3a4fd094ff49ef2485ab3f17d1e30b19103a70e4b24a7c3e2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/diffutils-3.10-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "diffutils-0__3.10-8.el10.x86_64",
    sha256 = "96882ec03cfc01ae557f0ec547fb8d346179eb705c899bec0533eafda7c1bd80",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/diffutils-3.10-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "dmidecode-1__3.6-5.el10.aarch64",
    sha256 = "381d5765cc5b1346f47dea4818c013bc308eb2cd9a76a9a3c4046a6982910956",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/dmidecode-3.6-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "dmidecode-1__3.6-5.el10.x86_64",
    sha256 = "332cfc77ea06aab27c93c1cf2382e50bf62ddad534c526795083a98ec10668c8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/dmidecode-3.6-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "duktape-0__2.7.0-10.el10.aarch64",
    sha256 = "c390a43273231fec4a25199690e0106268e3eb46a1592d4cd68cf56909efce5e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/duktape-2.7.0-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "duktape-0__2.7.0-10.el10.s390x",
    sha256 = "7cafae00eb1aa432b96c9fb9a6df9789d3ccf03515b7714c16ff8dcbaa7210d6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/duktape-2.7.0-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "duktape-0__2.7.0-10.el10.x86_64",
    sha256 = "23b7d2905723ed7adabe3362c54d54f0745c908029ec3be79bd881770d2c591a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/duktape-2.7.0-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "e2fsprogs-0__1.47.1-5.el10.aarch64",
    sha256 = "fd5592fb0e7c1ae9ae023eafb55c7ae3ac71c94c44e1f498f1eb56c1940f3c40",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/e2fsprogs-1.47.1-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "e2fsprogs-0__1.47.1-5.el10.s390x",
    sha256 = "23803262e02ed5ad895284267c828bee4620aa498326a36c659a36dcd12bce9e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/e2fsprogs-1.47.1-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "e2fsprogs-0__1.47.1-5.el10.x86_64",
    sha256 = "736291b66f30c8ad543f5bed5375c92bc8a2e3bce1704a77f5b727ee844fb0dd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/e2fsprogs-1.47.1-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "e2fsprogs-libs-0__1.47.1-5.el10.aarch64",
    sha256 = "e8b7d03d574363beaebef73048b8fe8461ed7b1206152b81eb0852f5c01d533b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/e2fsprogs-libs-1.47.1-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "e2fsprogs-libs-0__1.47.1-5.el10.s390x",
    sha256 = "25bb41764aefa735e891df10d2846b4c86f00f8eaabaf9a66acf08ebf290b700",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/e2fsprogs-libs-1.47.1-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "e2fsprogs-libs-0__1.47.1-5.el10.x86_64",
    sha256 = "d73c79a7bda1ce465707d82fa6b9777fcd2776301a6f6722ca323b4c9337c64b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/e2fsprogs-libs-1.47.1-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "edk2-aarch64-0__20251114-2.el10.aarch64",
    sha256 = "14b8a283058af0f4fb30f4e7c2235945b6420143a441ed31a3c1e976505ba2b8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/edk2-aarch64-20251114-2.el10.noarch.rpm",
    ],
)

rpm(
    name = "edk2-ovmf-0__20251114-2.el10.s390x",
    sha256 = "7568e5b29bb3644cceab0faf2d59578ec8de88f0dafad4808b91bc2680dee682",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/s390x/os/Packages/edk2-ovmf-20251114-2.el10.noarch.rpm",
    ],
)

rpm(
    name = "edk2-ovmf-0__20251114-2.el10.x86_64",
    sha256 = "7568e5b29bb3644cceab0faf2d59578ec8de88f0dafad4808b91bc2680dee682",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/edk2-ovmf-20251114-2.el10.noarch.rpm",
    ],
)

rpm(
    name = "elfutils-debuginfod-client-0__0.194-1.el10.aarch64",
    sha256 = "280b20ad99ef6a5097776c729d7b7ccc679d9eb4c977d32ee92af4641a8e745d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/elfutils-debuginfod-client-0.194-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "elfutils-debuginfod-client-0__0.194-1.el10.s390x",
    sha256 = "70da3d5d468afd29b27733d38e61b79ebeee2de0e75c5f11b9edbd5e151aa4fe",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/elfutils-debuginfod-client-0.194-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "elfutils-debuginfod-client-0__0.194-1.el10.x86_64",
    sha256 = "5ac0c4084d431eda2da1db7698d10d62195ec03f44e25755f4d6b8133d6606e6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/elfutils-debuginfod-client-0.194-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "elfutils-default-yama-scope-0__0.194-1.el10.aarch64",
    sha256 = "35f822daa4ecdce5dc624e6875d3b55491f8b5e0696d070672d2678036ad2ad0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/elfutils-default-yama-scope-0.194-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "elfutils-default-yama-scope-0__0.194-1.el10.s390x",
    sha256 = "35f822daa4ecdce5dc624e6875d3b55491f8b5e0696d070672d2678036ad2ad0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/elfutils-default-yama-scope-0.194-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "elfutils-default-yama-scope-0__0.194-1.el10.x86_64",
    sha256 = "35f822daa4ecdce5dc624e6875d3b55491f8b5e0696d070672d2678036ad2ad0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/elfutils-default-yama-scope-0.194-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "elfutils-libelf-0__0.194-1.el10.aarch64",
    sha256 = "97c0ad3cb708215214b2c79fce3e840eeb023e751a679c8da23b0ac24c9286b4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/elfutils-libelf-0.194-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "elfutils-libelf-0__0.194-1.el10.s390x",
    sha256 = "01795d511317f3717a7f837bf9e0ac92d5db4da33eb1fd5b93987313f6638fcf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/elfutils-libelf-0.194-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "elfutils-libelf-0__0.194-1.el10.x86_64",
    sha256 = "1bfacc8e5b007821e21f82b50aa1ab3f1a2959fd4f3361c277e75db43bd69284",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/elfutils-libelf-0.194-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "elfutils-libs-0__0.194-1.el10.aarch64",
    sha256 = "ca36cc469aae95470c33e08087f5176615ebe42453e06695c8897da87c8e6185",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/elfutils-libs-0.194-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "elfutils-libs-0__0.194-1.el10.s390x",
    sha256 = "f421c5e17662e93a3f0ba2d4511a206c4ee08c7bf2f7ee40e59d8c803c7c6097",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/elfutils-libs-0.194-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "elfutils-libs-0__0.194-1.el10.x86_64",
    sha256 = "6a6cce578a25f607ab0c593d889c9c52487f6c9019d3f9b4c3ebc2edb5dbbc89",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/elfutils-libs-0.194-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "expat-0__2.7.3-1.el10.aarch64",
    sha256 = "9d093b8a289a4fbac304097d8d628744fa0ea88f3a50a64c4ee1c657cb42a5c8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/expat-2.7.3-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "expat-0__2.7.3-1.el10.s390x",
    sha256 = "5fce4ab3c8a5e188f560bdbac6f780e36af2e71210f765153ee2c9328b8a2a5f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/expat-2.7.3-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "expat-0__2.7.3-1.el10.x86_64",
    sha256 = "e00c0876574daba5e70a3e2c86e21823fae1269b7a123d08ff5493a59dde3f36",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/expat-2.7.3-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "filesystem-0__3.18-17.el10.aarch64",
    sha256 = "6c4d8ecaf8b45c8d7d588c6ebe368a77805ed84830d0bc3b38e4c8e499514aba",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/filesystem-3.18-17.el10.aarch64.rpm",
    ],
)

rpm(
    name = "filesystem-0__3.18-17.el10.s390x",
    sha256 = "087e8def18ded2dd2a96f7a4292a3654704807d05f4424c43c0f5c873d7f9cb5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/filesystem-3.18-17.el10.s390x.rpm",
    ],
)

rpm(
    name = "filesystem-0__3.18-17.el10.x86_64",
    sha256 = "bcfb13f67c813d645f47e0a56d4bb76c0863deaf64ba93be8e0c30eecdc1e45e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/filesystem-3.18-17.el10.x86_64.rpm",
    ],
)

rpm(
    name = "findutils-1__4.10.0-5.el10.aarch64",
    sha256 = "f0e4db5b6e713c75e097e80218c592de4e6cb85d353f0933f64714df11b178b2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/findutils-4.10.0-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "findutils-1__4.10.0-5.el10.s390x",
    sha256 = "da20bdfeb9053ac3a1689d2ee2281298ee119175a8d486e4bb3eed1bc2857a94",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/findutils-4.10.0-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "findutils-1__4.10.0-5.el10.x86_64",
    sha256 = "c646c7c108a007d62792aa66e0bc9326312089a0f8bc1c9e9300b301fd2e4276",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/findutils-4.10.0-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "fips-provider-next-0__1.2.0-3.el10.aarch64",
    sha256 = "8b1a3f9bcf30fa7850ff5f068bb01b0c0b07385135a0155a50257f014f3156bd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/fips-provider-next-1.2.0-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "fips-provider-next-0__1.2.0-3.el10.s390x",
    sha256 = "f2d281204b6118905a9668f2cbc5ee832aa65cca751c3cfd6df5f94fd880c8ae",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/fips-provider-next-1.2.0-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "fips-provider-next-0__1.2.0-3.el10.x86_64",
    sha256 = "d1314bd57fd4e4bb2030519cd79ab562f8ce64866d51827cd4e0f73c190a6c9c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/fips-provider-next-1.2.0-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "fuse-0__2.9.9-25.el10.s390x",
    sha256 = "6d0dd7c5dc828fc93d96ff215d90324f8efd9e88a9512081f4cf6d6323387a2f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/fuse-2.9.9-25.el10.s390x.rpm",
    ],
)

rpm(
    name = "fuse-0__2.9.9-25.el10.x86_64",
    sha256 = "0707885f1d8074b5d36d85b4c60a68a10867894b379225302a94f3d54b6d4934",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/fuse-2.9.9-25.el10.x86_64.rpm",
    ],
)

rpm(
    name = "fuse-common-0__3.16.2-5.el10.s390x",
    sha256 = "86983857ec56f535e57283f302d9f344a348b55a9dc5e6e81ef388b397a14e2a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/fuse-common-3.16.2-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "fuse-common-0__3.16.2-5.el10.x86_64",
    sha256 = "eecc51472bf7713a97821ae02898b6811752aa513aa40dc5d380459fce590a40",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/fuse-common-3.16.2-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "fuse-libs-0__2.9.9-25.el10.s390x",
    sha256 = "65b86c79a139100f7d61acbef829a0a345c70316988cd7eb0f573f0c57e98647",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/fuse-libs-2.9.9-25.el10.s390x.rpm",
    ],
)

rpm(
    name = "fuse-libs-0__2.9.9-25.el10.x86_64",
    sha256 = "a8b094d60b9a7f83a84d8c7b0cdeed565be044dc2ecd170965b2c55ee4fa40f7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/fuse-libs-2.9.9-25.el10.x86_64.rpm",
    ],
)

rpm(
    name = "fuse3-libs-0__3.16.2-5.el10.aarch64",
    sha256 = "919f632731bc755d7c9c81d6faebb3bb703d7460ed72fdd65c453541d3999a72",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/fuse3-libs-3.16.2-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "fuse3-libs-0__3.16.2-5.el10.s390x",
    sha256 = "68501eaef0f538ca7e3731a4968f308ced9ae9c2a1b3b4d310890dd86b1843c5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/fuse3-libs-3.16.2-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "fuse3-libs-0__3.16.2-5.el10.x86_64",
    sha256 = "3482d8de135a306e94f7a35c1f8315b4e6acb699c1871ef28ddb02dc0fbdf7d6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/fuse3-libs-3.16.2-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gawk-0__5.3.0-6.el10.aarch64",
    sha256 = "16d7b639936dd4c8c977cd5b2ee3f5a02d3235954f67aa7485765a6b146683de",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gawk-5.3.0-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gawk-0__5.3.0-6.el10.s390x",
    sha256 = "0c918acd6aed7bbe461611db414bed4c1871b9ee9e4e5369460e016eb0c6bcbb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/gawk-5.3.0-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "gawk-0__5.3.0-6.el10.x86_64",
    sha256 = "ba59a3a4ee8741ed4e0c2517086164a76dc85309947f8b5ca7884f05c08ed959",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gawk-5.3.0-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gcc-0__14.3.1-4.3.el10.aarch64",
    sha256 = "0b3278b287510da35e3e4c01e91b8dd369c26ceec1666a056baf638a7a985169",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/gcc-14.3.1-4.3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gcc-0__14.3.1-4.3.el10.s390x",
    sha256 = "0d0a820dcc592e30a947859e38c5a285d33aa8567afde50908fe7b5cf556f30c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/gcc-14.3.1-4.3.el10.s390x.rpm",
    ],
)

rpm(
    name = "gcc-0__14.3.1-4.3.el10.x86_64",
    sha256 = "33e41378d8e45c67021c7a10d7a6ecf69836d68a1eab0412a8a0bb95475a0094",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/gcc-14.3.1-4.3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gdbm-1__1.23-14.el10.aarch64",
    sha256 = "0db16e24bf3d297cc3543842d63143f583de6ee157806b0a3dc51b5740a2722f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gdbm-1.23-14.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gdbm-1__1.23-14.el10.s390x",
    sha256 = "95c556f933af240938736727df962465928b7a556a8586b01e90c647facc2839",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/gdbm-1.23-14.el10.s390x.rpm",
    ],
)

rpm(
    name = "gdbm-1__1.23-14.el10.x86_64",
    sha256 = "159a6f1affc65d960c11a8726472699f693cec90a54a0862ad8340d0968f4838",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gdbm-1.23-14.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gdbm-libs-1__1.23-14.el10.aarch64",
    sha256 = "b46628d13eba77191aad6905de11fff87d6f45e52168e5b5365cb1f62078fd4d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gdbm-libs-1.23-14.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gdbm-libs-1__1.23-14.el10.s390x",
    sha256 = "38f1f8006c38c8fffa7f298bf3a143943f8611acaee7aad8200edc6bcde534aa",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/gdbm-libs-1.23-14.el10.s390x.rpm",
    ],
)

rpm(
    name = "gdbm-libs-1__1.23-14.el10.x86_64",
    sha256 = "b5f678293062eb1fcba572501d62e215dccfd222c26f5b76d9424f3c188cedee",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gdbm-libs-1.23-14.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gettext-0__0.22.5-6.el10.aarch64",
    sha256 = "27cba50dbb800aaf7f46bffa04003338c797472b334b08344b3633a60e0f1755",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gettext-0.22.5-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gettext-0__0.22.5-6.el10.s390x",
    sha256 = "02ab0b35769a517e0a2c255c4e4f23cfb9f661179355f81687a2d5b5198289d6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/gettext-0.22.5-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "gettext-0__0.22.5-6.el10.x86_64",
    sha256 = "19430ae2b77a7e4637bfcb70501748a27011f6c1e144a195b7046ecd9e6a96b4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gettext-0.22.5-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gettext-envsubst-0__0.22.5-6.el10.aarch64",
    sha256 = "ae3a179fff748702f7ad12bc2d8e58910d724a1d42f4cafb22af8ddfaf2eb216",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gettext-envsubst-0.22.5-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gettext-envsubst-0__0.22.5-6.el10.s390x",
    sha256 = "a9f2345a5875671c4d3a14ae491ea02b535d52ecb6ff65813aba392660d96065",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/gettext-envsubst-0.22.5-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "gettext-envsubst-0__0.22.5-6.el10.x86_64",
    sha256 = "f7b90e29f350fd67a2425a9d06c404371f1bbcdc43727452bf27c6c855d9eccf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gettext-envsubst-0.22.5-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gettext-libs-0__0.22.5-6.el10.aarch64",
    sha256 = "460e9216dbdd5a5a42bcd49162639e5515020a1caf9a246734ab7c19d5747b8e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gettext-libs-0.22.5-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gettext-libs-0__0.22.5-6.el10.s390x",
    sha256 = "aebeafee7bc7b3513b5210039214447304f4734e8ba4e590cbf74cfe0fb04393",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/gettext-libs-0.22.5-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "gettext-libs-0__0.22.5-6.el10.x86_64",
    sha256 = "de538283e9cc0281d53e05c235905a9e5c64ad1ac2533afb915ba75052f540a3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gettext-libs-0.22.5-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gettext-runtime-0__0.22.5-6.el10.aarch64",
    sha256 = "76d58cbcdddca202c4eecc30df7692d5f6e847f0ac233227349942b6f860a5da",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gettext-runtime-0.22.5-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gettext-runtime-0__0.22.5-6.el10.s390x",
    sha256 = "59a0988b7180c5b0c78c02b40c60f902f60d363beb3acc379c5d1ffd8fa6dfeb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/gettext-runtime-0.22.5-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "gettext-runtime-0__0.22.5-6.el10.x86_64",
    sha256 = "aec2ce3c3805190c65667c617e1ed100b65c251d16896819b0bc933ec3084ebf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gettext-runtime-0.22.5-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "glib2-0__2.80.4-11.el10.aarch64",
    sha256 = "34720321f5c846b69a1f9b36a928596dcadcf5e11c4d5298cc358c3fa184341d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/glib2-2.80.4-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "glib2-0__2.80.4-11.el10.s390x",
    sha256 = "8a5a41fd388e3f2b056effbdb5a88cc361650852679dd9e1942a85ff9e4d87d9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/glib2-2.80.4-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "glib2-0__2.80.4-11.el10.x86_64",
    sha256 = "8de6a8233b09fff4cd3acd29017ec6750af1bf5670772dcfa2b2bb80f6f85885",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/glib2-2.80.4-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "glibc-0__2.39-99.el10.aarch64",
    sha256 = "1812456204a03ba8f95f2a57d2dd2c150358b8a7702a4fbfb9065a91c6d150f1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/glibc-2.39-99.el10.aarch64.rpm",
    ],
)

rpm(
    name = "glibc-0__2.39-99.el10.s390x",
    sha256 = "dc8579cd4111606b3ee378b8bfbae86588f30ee083797cb60c843fd19829496d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/glibc-2.39-99.el10.s390x.rpm",
    ],
)

rpm(
    name = "glibc-0__2.39-99.el10.x86_64",
    sha256 = "cae4b2718a3a864f504056f35159be691f2091f3fa9a11b5a6e14ec1ed66f068",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/glibc-2.39-99.el10.x86_64.rpm",
    ],
)

rpm(
    name = "glibc-common-0__2.39-99.el10.aarch64",
    sha256 = "4cb977fcb1ee23b92bbc646a0a76370aa8ed6d2ec791f674be069e769b7e6b21",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/glibc-common-2.39-99.el10.aarch64.rpm",
    ],
)

rpm(
    name = "glibc-common-0__2.39-99.el10.s390x",
    sha256 = "644f0a1c2bc83994674bfe3f116799679e729c4d8484dd282c79270bc58ec9dd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/glibc-common-2.39-99.el10.s390x.rpm",
    ],
)

rpm(
    name = "glibc-common-0__2.39-99.el10.x86_64",
    sha256 = "48946309a086d9dd6f2dd6cf5ba0cba6560c6f29d6611990b609c043dcee2a0b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/glibc-common-2.39-99.el10.x86_64.rpm",
    ],
)

rpm(
    name = "glibc-devel-0__2.39-99.el10.aarch64",
    sha256 = "544ecb6fc79f28a455c809beb222bdca33c7806198885302ae529d66ff100818",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/glibc-devel-2.39-99.el10.aarch64.rpm",
    ],
)

rpm(
    name = "glibc-devel-0__2.39-99.el10.s390x",
    sha256 = "84c11de51a0e0ca7df13ae839b8f5cc9ab5c28ee151f1e78c8646b426e348d9a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/glibc-devel-2.39-99.el10.s390x.rpm",
    ],
)

rpm(
    name = "glibc-devel-0__2.39-99.el10.x86_64",
    sha256 = "b8b686c9cc6115aa8d0b658a1cc1676ebac0f225e49c4f2276b81874519f060b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/glibc-devel-2.39-99.el10.x86_64.rpm",
    ],
)

rpm(
    name = "glibc-langpack-et-0__2.39-99.el10.aarch64",
    sha256 = "72e9179aa2c0b865ca95d1a87b54d1e93ee4e3a2de62fa8bedab5ce43041d33b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/glibc-langpack-et-2.39-99.el10.aarch64.rpm",
    ],
)

rpm(
    name = "glibc-langpack-eu-0__2.39-99.el10.x86_64",
    sha256 = "b559ef68a9fc6a08ea3fe1cb0a66aeba6bce0b41cac70918ae90e39ca19c39cc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/glibc-langpack-eu-2.39-99.el10.x86_64.rpm",
    ],
)

rpm(
    name = "glibc-langpack-fo-0__2.39-99.el10.s390x",
    sha256 = "e78a7b207a4d9f16e33146a9a82b6c8a03307435bcc8c3f389d73671f7d0130d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/glibc-langpack-fo-2.39-99.el10.s390x.rpm",
    ],
)

rpm(
    name = "glibc-langpack-fy-0__2.39-99.el10.s390x",
    sha256 = "5e3aba50a85ce9a9f31ec65dec07157a610177481b90f5ab941dc504c744deeb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/glibc-langpack-fy-2.39-99.el10.s390x.rpm",
    ],
)

rpm(
    name = "glibc-minimal-langpack-0__2.39-99.el10.aarch64",
    sha256 = "04ef96e01570aab3aecc92c52a33e74042b76e5a531475b7e79bfc38723316e2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/glibc-minimal-langpack-2.39-99.el10.aarch64.rpm",
    ],
)

rpm(
    name = "glibc-minimal-langpack-0__2.39-99.el10.s390x",
    sha256 = "aa21d067dc02b52ed3a580711bc153da4399d95b5953683011c7e6be29f6f5ba",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/glibc-minimal-langpack-2.39-99.el10.s390x.rpm",
    ],
)

rpm(
    name = "glibc-minimal-langpack-0__2.39-99.el10.x86_64",
    sha256 = "0d0d2e2e8f233c4dd39a69ccdbfd8224ffa95fd44d3bc5a90328cbe2f5727b26",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/glibc-minimal-langpack-2.39-99.el10.x86_64.rpm",
    ],
)

rpm(
    name = "glibc-static-0__2.39-99.el10.aarch64",
    sha256 = "f35d8cde12d5c3f2ccda56905011442ce4c649543bb640332115ec4b7eb9f516",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/aarch64/os/Packages/glibc-static-2.39-99.el10.aarch64.rpm",
    ],
)

rpm(
    name = "glibc-static-0__2.39-99.el10.s390x",
    sha256 = "50c5864f59e62b246491eb1986e204f16355b068de80f8ecd416e0fc837e19d1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/s390x/os/Packages/glibc-static-2.39-99.el10.s390x.rpm",
    ],
)

rpm(
    name = "glibc-static-0__2.39-99.el10.x86_64",
    sha256 = "ee90381f83bf6d7804c7c0423e8c4a9c83e351f5754eac95d23ec0d4ce2fe9e1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/x86_64/os/Packages/glibc-static-2.39-99.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gmp-1__6.2.1-12.el10.aarch64",
    sha256 = "9bbe58df2a29320daf9b4c36305fcc7f781ab0bdd486736c6d8c685838141a41",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gmp-6.2.1-12.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gmp-1__6.2.1-12.el10.s390x",
    sha256 = "54d437788539933aa6de0963c6b1303e50b07f17db9ea847a71e19d1b4ef6a66",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/gmp-6.2.1-12.el10.s390x.rpm",
    ],
)

rpm(
    name = "gmp-1__6.2.1-12.el10.x86_64",
    sha256 = "6678824b5d45f9b66e8bfeb8f32736e0d710e3b38531a85548f55702d96b63a8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gmp-6.2.1-12.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gnupg2-0__2.4.5-3.el10.s390x",
    sha256 = "d1488ea4e7128106cbe7685ab5e608dc70d79fafde3c892b66506d038a5d0a67",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/gnupg2-2.4.5-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "gnupg2-0__2.4.5-3.el10.x86_64",
    sha256 = "32438ad3bc18d0d6d146ee6d2d97f34707e1391aa22d8f0d2eec203d67b082cb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gnupg2-2.4.5-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gnutls-0__3.8.10-2.el10.aarch64",
    sha256 = "34023920a6a73834417f61a1169fe8a3edda3acd1f0d780db037e8be01b3866f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gnutls-3.8.10-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gnutls-0__3.8.10-2.el10.s390x",
    sha256 = "08f1f6d00fd7513d03dc79f41595c3baad7f24171b3b6b8c671754e236999d85",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/gnutls-3.8.10-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "gnutls-0__3.8.10-2.el10.x86_64",
    sha256 = "0226b47f6900316b131298753165e14869c8e4eedce1d1819ae7a5e5b8bd9fac",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gnutls-3.8.10-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gnutls-dane-0__3.8.10-2.el10.aarch64",
    sha256 = "04cc91a776f6e19b6861f54d2d27e1ef856467eecde4658f8ac54d0aeb1bcb3c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/gnutls-dane-3.8.10-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gnutls-dane-0__3.8.10-2.el10.s390x",
    sha256 = "a66242f54e66cf76af97646047832eece06a47e80ef4f033c31d5f20053a3364",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/gnutls-dane-3.8.10-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "gnutls-dane-0__3.8.10-2.el10.x86_64",
    sha256 = "349be5a7e270c1d7a9e4c43d0a4c60d3408196d50a32261c0d1dd664a5a10954",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/gnutls-dane-3.8.10-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gnutls-utils-0__3.8.10-2.el10.aarch64",
    sha256 = "bf3239d5a05fe33a62da3aa79e027b10343be365040535c589c6d3c1a7e59dc0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/gnutls-utils-3.8.10-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gnutls-utils-0__3.8.10-2.el10.s390x",
    sha256 = "5c99377ef43bd1a5b641a78305b2f8cf10ccf27331562bb965dd8686637191c4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/gnutls-utils-3.8.10-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "gnutls-utils-0__3.8.10-2.el10.x86_64",
    sha256 = "c425c9f618245d385afd651d6cb72d84001a8e056bcb12ca3775f287b6349697",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/gnutls-utils-3.8.10-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gobject-introspection-0__1.79.1-6.el10.aarch64",
    sha256 = "a3bd85b169c321602bafe23ca724dfa2b897379a89384dfd453cbb3a03d25e66",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gobject-introspection-1.79.1-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gobject-introspection-0__1.79.1-6.el10.s390x",
    sha256 = "440ef891180126b7d295bca67df47a23bdf05dff3a43d535826b8aa82ad26bb3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/gobject-introspection-1.79.1-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "gobject-introspection-0__1.79.1-6.el10.x86_64",
    sha256 = "80913f97462db46c9962d539f325cef09bf85ab4c415a2c47b445fe96bba84b6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gobject-introspection-1.79.1-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "grep-0__3.11-10.el10.aarch64",
    sha256 = "d797740f7c738e5e7729949bde3d82274c5c6422242a82c1058fbe71ea0c37e9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/grep-3.11-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "grep-0__3.11-10.el10.s390x",
    sha256 = "d30a1ab1991131978b67f26d6c119f97bb5408a4bebac0294f2ac5417fe12276",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/grep-3.11-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "grep-0__3.11-10.el10.x86_64",
    sha256 = "a0eb701c640cd0a0c9195493a8fc9206fff62174d958ba4af2d92527191f803f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/grep-3.11-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "guestfs-tools-0__1.54.0-7.el10.s390x",
    sha256 = "a8afff6d24bfb91072d2a0c98ad7d574ac5840da0f5a97e92b717862f8f28492",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/guestfs-tools-1.54.0-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "guestfs-tools-0__1.54.0-7.el10.x86_64",
    sha256 = "8e050750f08fbb8a0fd2b0600e8d3acd966e74155bccafe0cd6eb0aaf9d087a3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/guestfs-tools-1.54.0-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "gzip-0__1.13-3.el10.aarch64",
    sha256 = "9b276d61a13e3c996f059a095881630fab9ec5a4a56a07ddc711e4db0a3362d4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/gzip-1.13-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "gzip-0__1.13-3.el10.s390x",
    sha256 = "d76be88d032b4f7525f5414d11081fb930fe338f108830b01b24f8501de3c2d5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/gzip-1.13-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "gzip-0__1.13-3.el10.x86_64",
    sha256 = "b7117230deceaba8bcd1341f0528df5855e54997cea04379fd3cc2c7c1e07ba8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/gzip-1.13-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "hexedit-0__1.6-8.el10.s390x",
    sha256 = "3b3fa64ec84f359ff667cf1c9f0c66e5300d08284d6a944e295fefbd3fd1e720",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/hexedit-1.6-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "hexedit-0__1.6-8.el10.x86_64",
    sha256 = "b4e61671ac71d0dc721f67b6a7d5ff28e3ec9c8e1b8251104bcd34d6f0611ce3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/hexedit-1.6-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "hivex-libs-0__1.3.24-2.el10.s390x",
    sha256 = "967989dfab46ed23e33b59c956d27ad881051582f1c59995824b93249c5ff004",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/hivex-libs-1.3.24-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "hivex-libs-0__1.3.24-2.el10.x86_64",
    sha256 = "ecf17b83680af8d8a3cef0a632ec3a7163d01c3dddd2364c9b47a4ba79e23150",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/hivex-libs-1.3.24-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "hwdata-0__0.379-10.6.el10.s390x",
    sha256 = "99183d83a278795a010aabf072072e4734ebfd27f33af0587b707e07017c54d4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/hwdata-0.379-10.6.el10.noarch.rpm",
    ],
)

rpm(
    name = "hwdata-0__0.379-10.6.el10.x86_64",
    sha256 = "99183d83a278795a010aabf072072e4734ebfd27f33af0587b707e07017c54d4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/hwdata-0.379-10.6.el10.noarch.rpm",
    ],
)

rpm(
    name = "iproute-0__6.17.0-1.el10.aarch64",
    sha256 = "44ddded795dfa336e7c553ee68f70d2ccfbe5f954849cfba4975d1a914565398",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/iproute-6.17.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "iproute-0__6.17.0-1.el10.s390x",
    sha256 = "902537a96edca3984a114f83b7a927b83316104ae3ba34c9606b4a964dea7aba",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/iproute-6.17.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "iproute-0__6.17.0-1.el10.x86_64",
    sha256 = "6ebcdb339cc28036f2dc26b8be2f38d628b7206c715c484dca6031631c304e88",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/iproute-6.17.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "iproute-tc-0__6.17.0-1.el10.aarch64",
    sha256 = "46b8ce7f1acdae7878bf667f40c63c3497576749ced449f51ba5b951ca9c39c8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/iproute-tc-6.17.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "iproute-tc-0__6.17.0-1.el10.s390x",
    sha256 = "50a10556d2351456e5e4cb6b5a17b608f731352c06046e4e5ce2bd7a6a387490",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/iproute-tc-6.17.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "iproute-tc-0__6.17.0-1.el10.x86_64",
    sha256 = "1956f5939d423ba743e5d30ca9310ed7e38c5b6cd671c966d0ed093edfa29ba9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/iproute-tc-6.17.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "iptables-libs-0__1.8.11-12.el10.aarch64",
    sha256 = "4bf894764ed0f9e7e92228587b8ec02962197b6ff87db3c0562081daf54efb40",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/iptables-libs-1.8.11-12.el10.aarch64.rpm",
    ],
)

rpm(
    name = "iptables-libs-0__1.8.11-12.el10.s390x",
    sha256 = "c94759e8d3245cfbe43eb96964ee9a4031585e6874535d564a00596cd56a4d76",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/iptables-libs-1.8.11-12.el10.s390x.rpm",
    ],
)

rpm(
    name = "iptables-libs-0__1.8.11-12.el10.x86_64",
    sha256 = "450dfc1d463564d4955c3a244cf190bd4544cf56288b9703e2d39af238494f6c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/iptables-libs-1.8.11-12.el10.x86_64.rpm",
    ],
)

rpm(
    name = "iputils-0__20240905-5.el10.aarch64",
    sha256 = "3ed67cca3fbb5f60f14f85ea712b1822f2a80c58287e744795ef995ebebc3761",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/iputils-20240905-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "iputils-0__20240905-5.el10.s390x",
    sha256 = "bf09f778f68c47515f0763e7c4aa952ed32dea57608a9473cb8edd50742e8a6a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/iputils-20240905-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "iputils-0__20240905-5.el10.x86_64",
    sha256 = "adfa1b26bf1cd23d0998c85da06ad787f0fa745bfd232f7acec225c1e88b05d8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/iputils-20240905-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "ipxe-roms-qemu-0__20240119-5.gitde8a0821.el10.x86_64",
    sha256 = "0b834df444ffe592d164f1dd5a2ce690417e459b8cd6d6c69b2075bbb9c8b4cb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/ipxe-roms-qemu-20240119-5.gitde8a0821.el10.noarch.rpm",
    ],
)

rpm(
    name = "jansson-0__2.14-3.el10.aarch64",
    sha256 = "a838d217420f9f10eb80a221b6cda50ff65e729c15be94f33cbb420f206ee880",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/jansson-2.14-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "jansson-0__2.14-3.el10.s390x",
    sha256 = "cc054f4efd4b779ec708061759de28acd9eb9df0ad8f3b32f9fe4752b1dcb06c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/jansson-2.14-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "jansson-0__2.14-3.el10.x86_64",
    sha256 = "25d2ef852d5941b27ae105ec780aa367605a6f8b86e6c6a13abdee1c1065979f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/jansson-2.14-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "json-c-0__0.18-3.el10.aarch64",
    sha256 = "d3ecfebff7515c94e971c9584b0815202712cc2642526ee4fe5e424ec8ff2fae",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/json-c-0.18-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "json-c-0__0.18-3.el10.s390x",
    sha256 = "d4bc7597af6496e70ffa04858c8d2418267b302959db9da3f89c6edfc723ccd4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/json-c-0.18-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "json-c-0__0.18-3.el10.x86_64",
    sha256 = "e73ae01d509fb9bef1bbd675be1c0003b0ee942a4187e9b14ef43e56e508245b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/json-c-0.18-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "json-glib-0__1.8.0-5.el10.aarch64",
    sha256 = "41de435cef6d704c1bb85066b9711e44f60b1dbff3574997094e5ed166e2b95e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/json-glib-1.8.0-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "json-glib-0__1.8.0-5.el10.s390x",
    sha256 = "851e663120bae993deed48bd36f06a84d9082b51b35c87244e7a8d3735bf422f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/json-glib-1.8.0-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "json-glib-0__1.8.0-5.el10.x86_64",
    sha256 = "156fddb0053ab256ec6ecbe7818c0ec8e957228eb2ed1d7cd244ecda85e1197e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/json-glib-1.8.0-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "kernel-headers-0__6.12.0-192.el10.aarch64",
    sha256 = "f69696aec37a229e1e427fad77d099027f44c8719c8c173b5ad99f7dc0cb7299",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/kernel-headers-6.12.0-192.el10.aarch64.rpm",
    ],
)

rpm(
    name = "kernel-headers-0__6.12.0-192.el10.s390x",
    sha256 = "155a3424d5df56088745abbfbecea8eccd2cd4ed99e2c3ce821bbb6eecfda661",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/kernel-headers-6.12.0-192.el10.s390x.rpm",
    ],
)

rpm(
    name = "kernel-headers-0__6.12.0-192.el10.x86_64",
    sha256 = "427895c58448d116f16c47b287fe5bd0de5984763390e7776bbd89431a1da1b2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/kernel-headers-6.12.0-192.el10.x86_64.rpm",
    ],
)

rpm(
    name = "keyutils-libs-0__1.6.3-5.el10.aarch64",
    sha256 = "a6ff394736256d5c2317ab5503a056d0f60155a92090853179506358bfd2333f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/keyutils-libs-1.6.3-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "keyutils-libs-0__1.6.3-5.el10.s390x",
    sha256 = "f2ed690c8ec6ef2a0b912ba324268d354d282f437b30044a44304543e78d9238",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/keyutils-libs-1.6.3-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "keyutils-libs-0__1.6.3-5.el10.x86_64",
    sha256 = "312e0bf42841bb330f7721012d1ee5816e5ea223e54fc5dfd1a95c6f1d7516b0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/keyutils-libs-1.6.3-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "kmod-0__31-12.el10.aarch64",
    sha256 = "2aa38be351be2c1b4efd0932928ebdc74217d23851f6b2aa0b86ce4fb3df2c12",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/kmod-31-12.el10.aarch64.rpm",
    ],
)

rpm(
    name = "kmod-0__31-12.el10.s390x",
    sha256 = "0273efd3eedba40d92e26ab6e72ff0dd27eba674c0f22ecf2e37a9d2dca1ab14",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/kmod-31-12.el10.s390x.rpm",
    ],
)

rpm(
    name = "kmod-0__31-12.el10.x86_64",
    sha256 = "9c08b22962d94f0d96cf156ec6e8624a537ab3821b52b4c0f725a7c059380d4e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/kmod-31-12.el10.x86_64.rpm",
    ],
)

rpm(
    name = "kmod-libs-0__31-12.el10.aarch64",
    sha256 = "3776b62bbebcd5862814723a738123b8903f2933c899e6221711e968600fc8f8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/kmod-libs-31-12.el10.aarch64.rpm",
    ],
)

rpm(
    name = "kmod-libs-0__31-12.el10.s390x",
    sha256 = "598b1bcbdd2a5806066fbd19eefc62efe53051acd09b530b5913f5dd2f151dcc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/kmod-libs-31-12.el10.s390x.rpm",
    ],
)

rpm(
    name = "kmod-libs-0__31-12.el10.x86_64",
    sha256 = "c1f5cd1f7bc9148f88ddac82d2c70b01a6ba871b7ef5cf81825d96c9ea2360de",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/kmod-libs-31-12.el10.x86_64.rpm",
    ],
)

rpm(
    name = "krb5-libs-0__1.21.3-8.el10.aarch64",
    sha256 = "ae0332e7dc9a151a1f86f44e8cd75148f8ced6aeb54d1d9671af752fd863b0c8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/krb5-libs-1.21.3-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "krb5-libs-0__1.21.3-8.el10.s390x",
    sha256 = "415688256b2cea553669441e078c28bcb9dc2227c4bbbd29c46c505cec994d6e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/krb5-libs-1.21.3-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "krb5-libs-0__1.21.3-8.el10.x86_64",
    sha256 = "c19429221a54c4de8c9d88d6c6e0f929d1c4199828300c6f94b667c6f7ff00f4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/krb5-libs-1.21.3-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "less-0__661-3.el10.s390x",
    sha256 = "6eb8705527ce26aa64dd692e6103e6b2197fb134a9e4f03e5db78d4c45035ddf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/less-661-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "less-0__661-3.el10.x86_64",
    sha256 = "1cf4afdf660772f65668cc702722facd7ed79d849150e14cab623cffd4167516",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/less-661-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libacl-0__2.3.2-4.el10.aarch64",
    sha256 = "20f3eeb53bf86dd2c7152fcdc33df3efd60777edd11f31f633739c0fdc0bdbf5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libacl-2.3.2-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libacl-0__2.3.2-4.el10.s390x",
    sha256 = "5f26a314b6e88e87516979610e9c0bda3dc55c67c9339bf79739574162eb1fa6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libacl-2.3.2-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "libacl-0__2.3.2-4.el10.x86_64",
    sha256 = "dd06cfe883fcdf7cb14b749180abfd9fe9924723341a8644a9f65c086febc647",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libacl-2.3.2-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libaio-0__0.3.111-22.el10.aarch64",
    sha256 = "99660f7b25fdb5503e0414e263ad91d0c1b61f1dc4e106721c0d1380b239d17f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libaio-0.3.111-22.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libaio-0__0.3.111-22.el10.s390x",
    sha256 = "5ee6ef6f4625016ae0746586e62fcf1c70596f8b557d8e6ad54cc67a4ae26690",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libaio-0.3.111-22.el10.s390x.rpm",
    ],
)

rpm(
    name = "libaio-0__0.3.111-22.el10.x86_64",
    sha256 = "ea807b22c77a37a766e62ad533dc3f9b80fd5b260016487cecea55b095092446",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libaio-0.3.111-22.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libarchive-0__3.7.7-4.el10.aarch64",
    sha256 = "c1a8850b22bb37325ee675db6c312615a4f3944777cd0cc3f24f72b13abc1ecb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libarchive-3.7.7-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libarchive-0__3.7.7-4.el10.s390x",
    sha256 = "6d1dedaf9597a0df9e485662dda1f971639912c40af3c7fbc1150dab8622900a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libarchive-3.7.7-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "libarchive-0__3.7.7-4.el10.x86_64",
    sha256 = "604bd62429638f12bd4699692e75f1bc0c2be2558c3e8abdf85d03974c443194",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libarchive-3.7.7-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libasan-0__14.3.1-4.3.el10.aarch64",
    sha256 = "a04297d8b681a9237a1034600a62e5fda0c394c496c2b3d88cdb0838a5788126",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libasan-14.3.1-4.3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libasan-0__14.3.1-4.3.el10.s390x",
    sha256 = "3384d53721a49340c07e02ca23a5bf138830c8c2bf1c6f8d144272abaf88c526",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libasan-14.3.1-4.3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libassuan-0__2.5.6-6.el10.s390x",
    sha256 = "d31b659dd6036b990ea71c10c426df13f2f395685ea8674fcca49d6a5fbeb580",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libassuan-2.5.6-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "libassuan-0__2.5.6-6.el10.x86_64",
    sha256 = "5cb1eff4efadf906bb8060bb41c205bf77eabe73142599053ae892ee7d85e9c0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libassuan-2.5.6-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libatomic-0__14.3.1-4.3.el10.aarch64",
    sha256 = "d069d8b89a5cf93b1ee185f3fb76a109084db9538b7fd0a90eaa5414699f43f6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libatomic-14.3.1-4.3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libatomic-0__14.3.1-4.3.el10.s390x",
    sha256 = "0596f69781d1d98bd986e6ed0f34ce2164789694fc590613817a83bf435e5b58",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libatomic-14.3.1-4.3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libattr-0__2.5.2-5.el10.aarch64",
    sha256 = "37a06ff130ff4112ca431839607e4d7c583ec4b0191431aa9913bba754880040",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libattr-2.5.2-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libattr-0__2.5.2-5.el10.s390x",
    sha256 = "583ef53e42b6928c6a707baee521a3161f0d00d094db96ce05b39ae4409c73f8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libattr-2.5.2-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "libattr-0__2.5.2-5.el10.x86_64",
    sha256 = "2ec3c5ba70aaae97db5226f07476c3fd0adfeed15d7cce3b676288273c829274",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libattr-2.5.2-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libblkid-0__2.40.2-15.el10.aarch64",
    sha256 = "255304ac6a0462e6cd059d128680f6c15dcd88cfa95cc7272f7387b17624a0c2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libblkid-2.40.2-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libblkid-0__2.40.2-15.el10.s390x",
    sha256 = "1cf3f37342b1b0cc20dace8837592749bc5cad8baaceb7f2b6480fb1ede7d895",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libblkid-2.40.2-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "libblkid-0__2.40.2-15.el10.x86_64",
    sha256 = "00cec7dfaf08b5ab015ab88bf41f263bab25d416993f271da6990f998eb7569c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libblkid-2.40.2-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libbpf-2__1.7.0-1.el10.aarch64",
    sha256 = "f89a67afcbc8eacc5c8c40e7c30ec5a5aaa78e89bca1dd1032b89c8634bbc605",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libbpf-1.7.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libbpf-2__1.7.0-1.el10.s390x",
    sha256 = "0d45b7988fb7d679b8f4f9e88439b086c0c086a1ba7853bfd376b5061ac3c12b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libbpf-1.7.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libbpf-2__1.7.0-1.el10.x86_64",
    sha256 = "1379b88512429975bbbdd65c8737cbc793664bef2f0e8c2e04c1481b939c85ca",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libbpf-1.7.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libbrotli-0__1.1.0-7.el10.s390x",
    sha256 = "6deea1eafedaa040d5c7c4af870f2ae2edb6742f02cd56cdf07789cf2acd1359",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libbrotli-1.1.0-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "libbrotli-0__1.1.0-7.el10.x86_64",
    sha256 = "9b397443a3ffe381380af22b55b0cac0f02412b859f92888138fba5e1df5e15d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libbrotli-1.1.0-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libburn-0__1.5.6-6.el10.aarch64",
    sha256 = "221ced92933bca63eb94d1ed60699f364e5d0b0b9915a0fba39d8b98d513d887",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libburn-1.5.6-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libburn-0__1.5.6-6.el10.s390x",
    sha256 = "6ac13f60b7e3bee622332411d4ec87c77dbd23f5710edf73209d29bf85904e4d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libburn-1.5.6-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "libburn-0__1.5.6-6.el10.x86_64",
    sha256 = "83ba66223a60bf93d13710b632ecc8c057c294a17f36ba31a18e16e7d4b97819",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libburn-1.5.6-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libcap-0__2.69-7.el10.aarch64",
    sha256 = "38c8ab1a8883b39bf46006ed39b7834cfa0df2ca0a4825908f7da4ed631c8fc6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libcap-2.69-7.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libcap-0__2.69-7.el10.s390x",
    sha256 = "3f7016928e759a177e4103d6c31dc0833ec43eae5836d871bc5ea540fd3d0c7b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libcap-2.69-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "libcap-0__2.69-7.el10.x86_64",
    sha256 = "54c14cb5c8dc3536f43d632d766ed302a8ecad2ad8efd6aa2d079dafc11d1cd9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libcap-2.69-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libcap-ng-0__0.8.4-6.el10.aarch64",
    sha256 = "993a88b692dbb7a73ec214c464d8c267155c87ebdb18fb3ecb8782a2f777ce31",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libcap-ng-0.8.4-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libcap-ng-0__0.8.4-6.el10.s390x",
    sha256 = "8256ca0cc7612b3bc6d2f86039c914f90c349fda27513cb42868242cb4949542",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libcap-ng-0.8.4-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "libcap-ng-0__0.8.4-6.el10.x86_64",
    sha256 = "38b2ce6018bc0c73cdbc79f5cb2bad63045d84c308a56085a9de4adfd3250add",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libcap-ng-0.8.4-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libcom_err-0__1.47.1-5.el10.aarch64",
    sha256 = "97380e5fe0fce42be70418333bae2d5d9044c5f7fbda30c9b28f7776718e76a2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libcom_err-1.47.1-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libcom_err-0__1.47.1-5.el10.s390x",
    sha256 = "09f3032ebafe4e8d93152c940f160b68df98b1a1c978fae86b8524a8d7713467",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libcom_err-1.47.1-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "libcom_err-0__1.47.1-5.el10.x86_64",
    sha256 = "37b036aa4cb44adade9c4206f2eea389035387082be0d2eda0249d9fd07fb842",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libcom_err-1.47.1-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libconfig-0__1.7.3-10.el10.s390x",
    sha256 = "2df21ddec6f917d330835cee60c9c71a604fa3f988aabe5f1deeebd76582dc5a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libconfig-1.7.3-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "libconfig-0__1.7.3-10.el10.x86_64",
    sha256 = "5bee52a5f0599fc6a59df28222e6e831c441887471412a3d18e6d13ddfaaa881",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libconfig-1.7.3-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libcurl-minimal-0__8.12.1-4.el10.aarch64",
    sha256 = "17a02035eee04463ae727468fb756a7088e83fe3db93138b8a1e5a6d8a7b2904",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libcurl-minimal-8.12.1-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libcurl-minimal-0__8.12.1-4.el10.s390x",
    sha256 = "4acfbb95ccb0ba29bd9598472942f3df90a6e3bbe7e4ec7893ec92dd97636a66",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libcurl-minimal-8.12.1-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "libcurl-minimal-0__8.12.1-4.el10.x86_64",
    sha256 = "053872b16ba35bdac16d7e3b3bed01fde3143face1ac9292dde9ac9b44a96758",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libcurl-minimal-8.12.1-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libeconf-0__0.6.2-4.el10.aarch64",
    sha256 = "1bb73420b4f72fb200ccc560224107e1ae62b8a7156051a88c9239c9def47983",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libeconf-0.6.2-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libeconf-0__0.6.2-4.el10.s390x",
    sha256 = "eb314cc56ffe80a641f97664b4d5a7313e3be7f1bef79e6636bcaab5751b350f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libeconf-0.6.2-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "libeconf-0__0.6.2-4.el10.x86_64",
    sha256 = "1cdb8e5bf4d7680e41ebb2b76da3aab34c1ece4bba2fed952d8f49da69117dfa",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libeconf-0.6.2-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libevent-0__2.1.12-16.el10.aarch64",
    sha256 = "275530b6896bec203e5cdf0cb427c78da43f5b01d3d26b0a2b239f2ad49fcda2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libevent-2.1.12-16.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libevent-0__2.1.12-16.el10.s390x",
    sha256 = "11e041d07e9f2f30736efa420ed437e11676bc796914f22c7b3fc322f08e997a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libevent-2.1.12-16.el10.s390x.rpm",
    ],
)

rpm(
    name = "libevent-0__2.1.12-16.el10.x86_64",
    sha256 = "f8f5c3946bbd53590978e9aeca3064d81ab580492c4ff8044c48797870276f47",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libevent-2.1.12-16.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libfdisk-0__2.40.2-15.el10.aarch64",
    sha256 = "946bd21e2a3b4dca038342bc02ced6e161415f3654bd117a6725510b5e44741c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libfdisk-2.40.2-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libfdisk-0__2.40.2-15.el10.s390x",
    sha256 = "db4c70cc90d26af06f4253703e436fe232e0ef0c954728eaa3501134f009cfb6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libfdisk-2.40.2-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "libfdisk-0__2.40.2-15.el10.x86_64",
    sha256 = "31aff408f90ee0628690b836477de4e5bf4ea3bc249fc4032085539a81480dbb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libfdisk-2.40.2-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libfdt-0__1.7.0-12.el10.aarch64",
    sha256 = "2a568810d2b8fbd8425eeb64738491a514feaec766f1b90d320caf320e543134",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libfdt-1.7.0-12.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libfdt-0__1.7.0-12.el10.s390x",
    sha256 = "bd49a0dac4411aca4319b6dc670e0fab0fb9672eaedd83aab40452878fbe46ac",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libfdt-1.7.0-12.el10.s390x.rpm",
    ],
)

rpm(
    name = "libfdt-0__1.7.0-12.el10.x86_64",
    sha256 = "9c519693ffe97be0dda3e06ea9708446b2692b5fd72b8cfb93f0b78ddc6418d5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libfdt-1.7.0-12.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libffi-0__3.4.4-10.el10.aarch64",
    sha256 = "87b620ad4069f0a9623913acc568a2659bcee3695293b275a9f09b809437bf6e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libffi-3.4.4-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libffi-0__3.4.4-10.el10.s390x",
    sha256 = "592be60a3f4ee70236cc254894f587012ce533a6f4fc74031bf6674792782338",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libffi-3.4.4-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "libffi-0__3.4.4-10.el10.x86_64",
    sha256 = "72aff2f3b4291f5418491e612be4f92d65a9239224a4906c0c63dbc4fc668e73",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libffi-3.4.4-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libgcc-0__14.3.1-4.3.el10.aarch64",
    sha256 = "f86c3466afd9a017bd0c9f7f26120b4ecc3e77eef0c0f6da7843765050012bc5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libgcc-14.3.1-4.3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libgcc-0__14.3.1-4.3.el10.s390x",
    sha256 = "e61bef933045b06a0b851d28c42556fdfedaed37aa704ee7b937ae890fe6627c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libgcc-14.3.1-4.3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libgcc-0__14.3.1-4.3.el10.x86_64",
    sha256 = "20a4555ff333952c625e39d5d0384161371a1f8fb037f0f1c8800ed3778abd14",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libgcc-14.3.1-4.3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libgcrypt-0__1.11.0-6.el10.s390x",
    sha256 = "b4383e8187d076c47b732487866eeebc0e40c1474e0dd6d927dfca6172cb274f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libgcrypt-1.11.0-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "libgcrypt-0__1.11.0-6.el10.x86_64",
    sha256 = "1be7cfbc9f69f9e2b3d3f0621e14ded96e27d1c334decb5c88d1e396edf825e2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libgcrypt-1.11.0-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libgomp-0__14.3.1-4.3.el10.aarch64",
    sha256 = "741baaea475d1bbd2078e5366be928caa51f476e4cb2b2ca7d45ea7523ca1d8f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libgomp-14.3.1-4.3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libgomp-0__14.3.1-4.3.el10.s390x",
    sha256 = "9f561bcd63e706eb8b6cc6253223ef66ac2d995d640413b76de2080dad05dddf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libgomp-14.3.1-4.3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libgomp-0__14.3.1-4.3.el10.x86_64",
    sha256 = "bfdcba2fd598203e366fb8379c8d76442e9dd5763a4c60ce42af0bf712a6df8c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libgomp-14.3.1-4.3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libgpg-error-0__1.50-2.el10.s390x",
    sha256 = "d2ae277868010dcf3003a59234b70743f2ff0846e0f8aba6cf349de19d6c2173",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libgpg-error-1.50-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libgpg-error-0__1.50-2.el10.x86_64",
    sha256 = "b7d74c79f82abf581fdb5b9fbd0b3792640c26780652036be284347b7b339fff",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libgpg-error-1.50-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libguestfs-1__1.58.1-2.el10.s390x",
    sha256 = "5546a48cbc10fe6dac94e219d9458be67cfb047a60482e4429987461a4f7ee71",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libguestfs-1.58.1-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libguestfs-1__1.58.1-2.el10.x86_64",
    sha256 = "9e258c41aeb346fd4e37d13a3346f89c9a5d1354586da91567a948e520da78bd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libguestfs-1.58.1-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libibverbs-0__57.0-2.el10.aarch64",
    sha256 = "c542dc9c95c8a74c8521b64c971fa2a9415ee78becb5ec22dd8fc991be9d36f1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libibverbs-57.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libibverbs-0__57.0-2.el10.s390x",
    sha256 = "ace62b9368b6f207ecc03642baadac205ab4fc4247da37530dbf3e3b77a1216f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libibverbs-57.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libibverbs-0__57.0-2.el10.x86_64",
    sha256 = "370284438b7e09f12e250dfb345c3d6a57263404009ebab56bc8271d0650ff19",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libibverbs-57.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libidn2-0__2.3.7-3.el10.aarch64",
    sha256 = "947248aeedd08f88d9490f3020dee6416595cf8d25e15738c306d55e9cae8bfc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libidn2-2.3.7-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libidn2-0__2.3.7-3.el10.s390x",
    sha256 = "7d75542211c9b7b8e53aaf6aee0dc430f091414fb2ca755baadc35f844bded75",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libidn2-2.3.7-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libidn2-0__2.3.7-3.el10.x86_64",
    sha256 = "04ae61bbe2cc0db7581d6f96a562b9b87a8a4dba714a0cf2c73bba6306e94c27",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libidn2-2.3.7-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libisoburn-0__1.5.6-6.el10.aarch64",
    sha256 = "038aa1a45c117b4c4abbeedc3f67faa01a66570bc554f0e53955ce83a26e7281",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libisoburn-1.5.6-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libisoburn-0__1.5.6-6.el10.s390x",
    sha256 = "52fa5b55814330bf00ce834de3fa86c0dff29691feb242800f04cc13c44f2f3a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libisoburn-1.5.6-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "libisoburn-0__1.5.6-6.el10.x86_64",
    sha256 = "6f41c5e8e0d9dedf3c0b07f2ddc0748870cb18ed15a23dc22038a022c1d38e74",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libisoburn-1.5.6-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libisofs-0__1.5.6-6.el10.aarch64",
    sha256 = "cacea726d7dd126a364ec2431f49417b884287c84664042d51d59011acac34d6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libisofs-1.5.6-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libisofs-0__1.5.6-6.el10.s390x",
    sha256 = "655af71fa634d1bb1c86b6c4810c452c98c1772dc9d55da3e3e9f2413bafc293",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libisofs-1.5.6-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "libisofs-0__1.5.6-6.el10.x86_64",
    sha256 = "a61eda57352e86657ca99fc774d0d74c0fbd67dad5bcd02139ff6de466a038ac",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libisofs-1.5.6-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libksba-0__1.6.7-2.el10.s390x",
    sha256 = "f4a0e294968ce54ad30e4de5baabcdebd7a9db7900266e4feb7cadecd7f18cba",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libksba-1.6.7-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libksba-0__1.6.7-2.el10.x86_64",
    sha256 = "2bfa8330ad9c63eaecd2bd1d0989625e812d853a85c505cb759a2d1c06750607",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libksba-1.6.7-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libmnl-0__1.0.5-7.el10.aarch64",
    sha256 = "786a9caea9de8f4529e5ec07ac24c9cfccd50ee5f4b045c37dbd4eb074b34f34",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libmnl-1.0.5-7.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libmnl-0__1.0.5-7.el10.s390x",
    sha256 = "bb6229f5c62e69828f94cd16f0086115e17ea9c9ab831ad14781f9b9807cd3d8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libmnl-1.0.5-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "libmnl-0__1.0.5-7.el10.x86_64",
    sha256 = "1e1d36725d958fc3f9016cc85b238bf5462a43ae7be144d0a4a72be0982d68ce",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libmnl-1.0.5-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libmount-0__2.40.2-15.el10.aarch64",
    sha256 = "5d5edf37b93295b534e1ce55f4a7370e08f6831b1b9ea3d18085a889ea7dd111",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libmount-2.40.2-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libmount-0__2.40.2-15.el10.s390x",
    sha256 = "652be9673bbc71ea2c71e0c7295dfbba5f39c3e222a0412cdf8b07d66b09ce6d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libmount-2.40.2-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "libmount-0__2.40.2-15.el10.x86_64",
    sha256 = "857a25a9634578ee103810bf684d3ec0c881c258d9e74240ff77962a70a2e6a2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libmount-2.40.2-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libmpc-0__1.3.1-7.el10.aarch64",
    sha256 = "bb46a7465559a26c085bf1c02f0764332430a6c1b8fb3f08c8cee184e3d1f02a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libmpc-1.3.1-7.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libmpc-0__1.3.1-7.el10.s390x",
    sha256 = "ad956e3c217ba500101acb4219f4e07390ca5ac8a14f99ca9cca85220b525da1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libmpc-1.3.1-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "libmpc-0__1.3.1-7.el10.x86_64",
    sha256 = "daaa73a35dfe21a8201581e333b79ccd296ae87a93f9796ba522e58edc23777c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libmpc-1.3.1-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libnbd-0__1.24.0-1.el10.aarch64",
    sha256 = "393fb9a22ff850b0e2b2523bc7ac553af21c11e8c1c2f9c25ac5230526fdf490",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libnbd-1.24.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libnbd-0__1.24.0-1.el10.s390x",
    sha256 = "77090bbbd3ea707fce4ebbe8949f4bf594608de81675630ee51b7e1199c82cd8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libnbd-1.24.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libnbd-0__1.24.0-1.el10.x86_64",
    sha256 = "0b44c10ebfb4be2c12275c1edf9063ee45624edea98d01c76fc1a0ec61723ffd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libnbd-1.24.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libnetfilter_conntrack-0__1.0.9-12.el10.aarch64",
    sha256 = "53c1b45e66ef040f6486052395cfda198d9b8b3058834ae7e8b7864b04f9c766",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libnetfilter_conntrack-1.0.9-12.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libnetfilter_conntrack-0__1.0.9-12.el10.s390x",
    sha256 = "f7f15dc88c380db9cbf6e62c1a04b872867d7f5a4f2cc0b2f9988db963d8401d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libnetfilter_conntrack-1.0.9-12.el10.s390x.rpm",
    ],
)

rpm(
    name = "libnetfilter_conntrack-0__1.0.9-12.el10.x86_64",
    sha256 = "71af0b9fb8b790e3d471a74ef463dc3cbb0267c9bdbaf876160fa9821f63200f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libnetfilter_conntrack-1.0.9-12.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libnfnetlink-0__1.0.2-3.el10.aarch64",
    sha256 = "0ac9c6ebd2c5652bea632435fd73bfb36785d2f15eb840ca53049d6bb3abe639",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libnfnetlink-1.0.2-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libnfnetlink-0__1.0.2-3.el10.s390x",
    sha256 = "433915ab5525daf404ae37d47f1f735d257a33ff8a2565741e3449d5b39b0cab",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libnfnetlink-1.0.2-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libnfnetlink-0__1.0.2-3.el10.x86_64",
    sha256 = "2988f90762058160e4071b79b96523901a2170bc6242488878ce51fbc0d871ca",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libnfnetlink-1.0.2-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libnftnl-0__1.3.0-2.el10.aarch64",
    sha256 = "c420a00ab70913201a953ed39ea017980c0b9a92132b93700a53bfe7c8d04bfe",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libnftnl-1.3.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libnftnl-0__1.3.0-2.el10.s390x",
    sha256 = "f2a520590d440b4aa79cbf03b42afa60b7918751917df8b9f55cd77924637f86",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libnftnl-1.3.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libnftnl-0__1.3.0-2.el10.x86_64",
    sha256 = "3733a952b42041ea2f83eb8ce39000ab4ea713008a9ac1a22e2f13ef10eab93e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libnftnl-1.3.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libnghttp2-0__1.64.0-2.el10.aarch64",
    sha256 = "3bb538521491eaf5ee76c9bce5e0668345f6c03c5eb8610375ea62fc13252cf5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libnghttp2-1.64.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libnghttp2-0__1.64.0-2.el10.s390x",
    sha256 = "2731569708badb4582ed74326175366aca1e1a86a90fcec460acaefdb6c4543f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libnghttp2-1.64.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libnghttp2-0__1.64.0-2.el10.x86_64",
    sha256 = "087a6ea4e234b3a6b12326f5da756c8010efdda58a7a8ebcbf4f4da32247a566",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libnghttp2-1.64.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libnl3-0__3.11.0-1.el10.aarch64",
    sha256 = "b27497d441cd6ae6fbf6a077913eff334c64ace7b6030d6c11a221316a8c8d92",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libnl3-3.11.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libnl3-0__3.11.0-1.el10.s390x",
    sha256 = "4171ef398ec29504e828c461581fa44e6afe6b36ff84da66f902edd932271b34",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libnl3-3.11.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libnl3-0__3.11.0-1.el10.x86_64",
    sha256 = "886324f9d4b8c95a46d5c77c0f5cd90051c83462e78ca752c70cc709f5a01d90",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libnl3-3.11.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libosinfo-0__1.11.0-8.el10.s390x",
    sha256 = "3ae1d93313b6b63d6cf2887307a539e9a371d64134ad7a363d5b31c64ee2734d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libosinfo-1.11.0-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "libosinfo-0__1.11.0-8.el10.x86_64",
    sha256 = "e632610473056869e88ddaa33511e043721d89f8c2216590a274e638f64e98fa",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libosinfo-1.11.0-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libpcap-14__1.10.4-7.el10.aarch64",
    sha256 = "f15a71822ba8269643911bdd5455e2c24c2489927217d3512c9453a6ff8af5bf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libpcap-1.10.4-7.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libpcap-14__1.10.4-7.el10.s390x",
    sha256 = "98f721f0b8b731a77d27bcb3178c7297e84372095a452ab52a98192c76953d3f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libpcap-1.10.4-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "libpcap-14__1.10.4-7.el10.x86_64",
    sha256 = "a933eb7fba1535c9df52f7e44504535b25f8b6fb79c5cf68a0b6e80eb4b9dbf8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libpcap-1.10.4-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libpkgconf-0__2.1.0-3.el10.aarch64",
    sha256 = "fc2db71a801f4cd03425463d0aea745da36837f25d8cc2042eb747c8a336f989",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libpkgconf-2.1.0-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libpkgconf-0__2.1.0-3.el10.s390x",
    sha256 = "7d503fcdd8154231531ec1e076ac2552b9d0a5fe096fb50d3a9ff0ebce07d92d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libpkgconf-2.1.0-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libpkgconf-0__2.1.0-3.el10.x86_64",
    sha256 = "813f59114413d5e14fc566262ee3d4b56b621beacbe40eda6f28d31f464de1a6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libpkgconf-2.1.0-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libpng-2__1.6.40-8.el10.aarch64",
    sha256 = "449443264f27154b3453af2deb2bb91ab184994b8e2cc94f45d93aa56381c081",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libpng-1.6.40-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libpng-2__1.6.40-8.el10.s390x",
    sha256 = "434d4e41e3409a7907219176b8b366f3f9c9c761a152af011a4d794c4972e1ab",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libpng-1.6.40-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "libpng-2__1.6.40-8.el10.x86_64",
    sha256 = "215a0ac1a843c31f8d77b77c03719ecf96d5033ec83e95000c8f5f669bfbca95",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libpng-1.6.40-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libpsl-0__0.21.5-6.el10.s390x",
    sha256 = "ef89d923a60bc9658e5524a80960a865d805aa136b7dd3761a162d58b2aff46d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libpsl-0.21.5-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "libpsl-0__0.21.5-6.el10.x86_64",
    sha256 = "1dca94a85aabd9730bc731fa8a6abb138fec28b75c6a39694d862135c2ade0f3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libpsl-0.21.5-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libpwquality-0__1.4.5-12.el10.aarch64",
    sha256 = "0d0d6a0e741f94889796b551935f72cf551587067f0c9b64531b5c34b03ab1d8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libpwquality-1.4.5-12.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libpwquality-0__1.4.5-12.el10.s390x",
    sha256 = "bff94322487bd0bd36640c27e56d7b0167187772cab630758bc56aada0038aea",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libpwquality-1.4.5-12.el10.s390x.rpm",
    ],
)

rpm(
    name = "libpwquality-0__1.4.5-12.el10.x86_64",
    sha256 = "eda9e6acc99c2c9fa058a9db428da1b0c7441f2be174b9aa7f1628359e36e6ab",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libpwquality-1.4.5-12.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libseccomp-0__2.5.6-1.el10.aarch64",
    sha256 = "322aa4ea140a63645c7f086b58a08346617eea2efee8044287b76373d633b65f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libseccomp-2.5.6-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libseccomp-0__2.5.6-1.el10.s390x",
    sha256 = "e652de14f8c0d52480c2bba779daf5da7b7fd66e65090e1781f41d3bee250840",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libseccomp-2.5.6-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libseccomp-0__2.5.6-1.el10.x86_64",
    sha256 = "654051862cc301ed43501ab36b687ed5adeb3ca57689f54a80bf760ad9686e54",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libseccomp-2.5.6-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libselinux-0__3.9-3.el10.aarch64",
    sha256 = "9df151baa8c60a2bc4998a636b7f50b40bf86d64420578de7501a9daf1d25a31",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libselinux-3.9-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libselinux-0__3.9-3.el10.s390x",
    sha256 = "9b95b6d120c78041499f330aeb8092bea16fd01536f518445618034599741182",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libselinux-3.9-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libselinux-0__3.9-3.el10.x86_64",
    sha256 = "9030f7855d93e37b4d2d9e1d7a05522e059f6f90f328ad9433125eedd21f35fa",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libselinux-3.9-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libselinux-utils-0__3.9-3.el10.aarch64",
    sha256 = "971a828ba86b404b4d689dcae7725c62a08bfc7f4c460e5e9a80ff2c3428ea78",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libselinux-utils-3.9-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libselinux-utils-0__3.9-3.el10.s390x",
    sha256 = "b61a38bf43fe870be78dabd06af7b99f68f450a14fbdc7590ecfcd706d435617",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libselinux-utils-3.9-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libselinux-utils-0__3.9-3.el10.x86_64",
    sha256 = "794445213c267c95ed897828e76671463e6dd8fe2b4003e6ad5f4ef64437a5e6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libselinux-utils-3.9-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libsemanage-0__3.9-2.el10.aarch64",
    sha256 = "cc0aab6cace2891a9991af486461d5507c26b352e7f335e5558404aa3a2d23ae",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libsemanage-3.9-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libsemanage-0__3.9-2.el10.s390x",
    sha256 = "773fce11f06bb13d5e7c821ef80d478e07ef277ab08272e0498b05a3f380c53c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libsemanage-3.9-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libsemanage-0__3.9-2.el10.x86_64",
    sha256 = "20d1fd8d8e72ab3efd493fd6c24cf134bdcb351d9f235a6a479f10097bc6f516",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libsemanage-3.9-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libsepol-0__3.9-1.el10.aarch64",
    sha256 = "7aadc40ad02a2d595d01e6322a005bd68278b905b95aa8b517cff89a8a3cbf21",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libsepol-3.9-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libsepol-0__3.9-1.el10.s390x",
    sha256 = "436ef10d6dc7d82c06620e43ae78edb153a1d9bd16da55786be538a0560669b5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libsepol-3.9-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libsepol-0__3.9-1.el10.x86_64",
    sha256 = "3bd100da5da32dc933544f2206043a0fc34c94a5d249a3129dc9305e1844df0c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libsepol-3.9-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libslirp-0__4.7.0-10.el10.aarch64",
    sha256 = "077e56fc67d139c2569bdd6b920777df742773a155425b11401406aac39f4e7b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libslirp-4.7.0-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libslirp-0__4.7.0-10.el10.s390x",
    sha256 = "392067b3525f2d603a121d6a2b7e5683c7a903a7be677ffc94fe2f1b278d3a11",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libslirp-4.7.0-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "libslirp-0__4.7.0-10.el10.x86_64",
    sha256 = "bc98bf4c15d226b809474c1237700e4e3158d77c4b7488611599672ac0b570af",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libslirp-4.7.0-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libsmartcols-0__2.40.2-15.el10.aarch64",
    sha256 = "803ad2645c105834a9ce3b89cd74653ba9278ac28cf25e3a6fd0bb60d08cdd11",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libsmartcols-2.40.2-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libsmartcols-0__2.40.2-15.el10.s390x",
    sha256 = "ea2a641dc725f5fa0ac7d5a1fc5f564f41e7f806214941175e7ceec14f36bb8a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libsmartcols-2.40.2-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "libsmartcols-0__2.40.2-15.el10.x86_64",
    sha256 = "18ae7ebcafe3fa1f0d7b1bc9290a8cf6e087da36a82c951031d00e0457d46859",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libsmartcols-2.40.2-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libsoup3-0__3.6.5-5.el10.s390x",
    sha256 = "eb1082bb8403619c3ce9352feab119dd27eaf3e10417d7968c0342c0269d6272",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libsoup3-3.6.5-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "libsoup3-0__3.6.5-5.el10.x86_64",
    sha256 = "38b4b20b159ae75afc780d10b1f212a2337a750e42c9183b071f5164a6cdfeac",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libsoup3-3.6.5-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libss-0__1.47.1-5.el10.aarch64",
    sha256 = "edb0a7af06913af8ce5a72ab62de780b8b08ad7a7db6cabc4bae9b95d4253607",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libss-1.47.1-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libss-0__1.47.1-5.el10.s390x",
    sha256 = "1a38ebb62511e39bae10e1cbde35152562e90d8014f243696bed5c3deef6bf9f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libss-1.47.1-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "libss-0__1.47.1-5.el10.x86_64",
    sha256 = "2dea843b06f0bd161807d8cfd7e3ef05f5944f5ea57332e5b58700bd93262765",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libss-1.47.1-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libssh-0__0.11.1-3.el10.aarch64",
    sha256 = "95144fadc027f326413a7960e36ca57fd72c94eb76692632a3cb1184eea19e81",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libssh-0.11.1-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libssh-0__0.11.1-3.el10.s390x",
    sha256 = "7fc02fabb13f1c1fe5322e99bf27f9b683361b1159de6f83cec27662524e51ab",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libssh-0.11.1-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libssh-0__0.11.1-3.el10.x86_64",
    sha256 = "ef979981656e3422ce8ab958097db269e235f5f17eb9d58aeb5c9c41a6844d97",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libssh-0.11.1-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libssh-config-0__0.11.1-3.el10.aarch64",
    sha256 = "a894c61e02dfb1d9630deb0cada5c1508bdfe63ea0bd906fab9c8bb0f5a0418d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libssh-config-0.11.1-3.el10.noarch.rpm",
    ],
)

rpm(
    name = "libssh-config-0__0.11.1-3.el10.s390x",
    sha256 = "a894c61e02dfb1d9630deb0cada5c1508bdfe63ea0bd906fab9c8bb0f5a0418d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libssh-config-0.11.1-3.el10.noarch.rpm",
    ],
)

rpm(
    name = "libssh-config-0__0.11.1-3.el10.x86_64",
    sha256 = "a894c61e02dfb1d9630deb0cada5c1508bdfe63ea0bd906fab9c8bb0f5a0418d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libssh-config-0.11.1-3.el10.noarch.rpm",
    ],
)

rpm(
    name = "libsss_idmap-0__2.12.0-1.el10.aarch64",
    sha256 = "5d56064a3fe65a8eb6b9e02b239587bc9a3e132a947d836c5edf4f02d538fd03",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libsss_idmap-2.12.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libsss_idmap-0__2.12.0-1.el10.s390x",
    sha256 = "573803676a86b4f7d8ed234ce70c8a40412249a4bdbd8256c4fac09fb52eb2c9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libsss_idmap-2.12.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libsss_idmap-0__2.12.0-1.el10.x86_64",
    sha256 = "fbfda0b5f6b73eb4b75196370ea91361afb736e08212f291f6c5328ecab401de",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libsss_idmap-2.12.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libsss_nss_idmap-0__2.12.0-1.el10.aarch64",
    sha256 = "fff9355f09c72d9e9f6eddf6a9b5ae5d91d03d34c9e09acd4790af60ef7e2fb5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libsss_nss_idmap-2.12.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libsss_nss_idmap-0__2.12.0-1.el10.s390x",
    sha256 = "cd8c9d8edca4e2658839230d7afd66c8e5c52c782169eb1cd1354b8683f9d606",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libsss_nss_idmap-2.12.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libsss_nss_idmap-0__2.12.0-1.el10.x86_64",
    sha256 = "690e064654f7f3ede0eed89565408970258329a37580c61ee04872a64c489b32",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libsss_nss_idmap-2.12.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libstdc__plus____plus__-0__14.3.1-4.3.el10.aarch64",
    sha256 = "150b07a914aaf05801ebb3a9cb539c8291003738a5b0a86cb44ad07a8a99032c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libstdc++-14.3.1-4.3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libstdc__plus____plus__-0__14.3.1-4.3.el10.s390x",
    sha256 = "5d0f9667179fce3828c81c8fe2e2806de7841222d25251dd4371acb72966ec77",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libstdc++-14.3.1-4.3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libstdc__plus____plus__-0__14.3.1-4.3.el10.x86_64",
    sha256 = "c58029947e33553509b21496b23445b7c85c9ca2dd3013bd7f2c8a25a32de12b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libstdc++-14.3.1-4.3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libtasn1-0__4.20.0-1.el10.aarch64",
    sha256 = "f46e93f5bff81ef89c872c2ad91ddde57c9ee0025d162647618ba5e764520854",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libtasn1-4.20.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libtasn1-0__4.20.0-1.el10.s390x",
    sha256 = "1c866239a4d6d0198fb9916c5ae132f19ccc576cb389101270291e7c1b5e3f1d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libtasn1-4.20.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libtasn1-0__4.20.0-1.el10.x86_64",
    sha256 = "6f88995a1e9181e8d99b77b8cc60681f79ba424382a17e590c3d813f300adf65",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libtasn1-4.20.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libtirpc-0__1.3.5-1.el10.aarch64",
    sha256 = "6e0345c38ef8c15d2f1743892063241f0273a6b14c1844ab127ad0c085b510e1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libtirpc-1.3.5-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libtirpc-0__1.3.5-1.el10.s390x",
    sha256 = "5f6160af1ea75ef4df15281225e50e13d7de69aad58a7bc08d558ffb88c86086",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libtirpc-1.3.5-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "libtirpc-0__1.3.5-1.el10.x86_64",
    sha256 = "8692d388ed8b7fa6ffe56c9403576ea7b49d55c305e9a64dd44a59fb592fe295",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libtirpc-1.3.5-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libtpms-0__0.9.6-11.el10.aarch64",
    sha256 = "3c666376aabf7fa14a76232e8709a390587bcebfebb24897de6c7c693703fed0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libtpms-0.9.6-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libtpms-0__0.9.6-11.el10.s390x",
    sha256 = "55e810e2e6a3c8b166c1fa48a38e2430a2f3b1587014009f48b88fc17ba2f5c4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libtpms-0.9.6-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "libtpms-0__0.9.6-11.el10.x86_64",
    sha256 = "595a554e74b9e9515d2d615b05ed13118f4a9e0c21f1afc74bf5b4e677b56c9a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libtpms-0.9.6-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libubsan-0__14.3.1-4.3.el10.aarch64",
    sha256 = "36fb1bf1480e59bfe4a5ed8aedf5c10323bc8c0410b7ffd36721ec6ebafbebb7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libubsan-14.3.1-4.3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libubsan-0__14.3.1-4.3.el10.s390x",
    sha256 = "f44ed900f63cc5709b5798190a644045def5d44464db671176005ebcd6618372",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libubsan-14.3.1-4.3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libunistring-0__1.1-10.el10.aarch64",
    sha256 = "aa793b61f51cb8727c37520bc4b261845831b9a5789649a798c4e8a2cc207f4f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libunistring-1.1-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libunistring-0__1.1-10.el10.s390x",
    sha256 = "9c45ddec6ffa51201a570b9881e9277b8f22c8eb40ee62b45d3c2b86bdc8eeac",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libunistring-1.1-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "libunistring-0__1.1-10.el10.x86_64",
    sha256 = "603c06593a43f5766a53588d9ba18855ddb7b238963b8e09d8a328a17959b774",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libunistring-1.1-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "liburing-0__2.12-1.el10.aarch64",
    sha256 = "29f16a2950ef7ddaae31d7806d98961bf1f7d1772623782ae45cc687a3980c62",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/liburing-2.12-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "liburing-0__2.12-1.el10.s390x",
    sha256 = "53f3015878c4044a13caf8060a4afa6c10aff93452aa4b8de84cd2374d456a51",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/liburing-2.12-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "liburing-0__2.12-1.el10.x86_64",
    sha256 = "133309fc854ab7859713d7944e5a14e8cbc3f3916bbcd9f9e6af4d4850424c15",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/liburing-2.12-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libusb1-0__1.0.29-3.el10.aarch64",
    sha256 = "e0d9019535c50ac90e39e158bcf1b4ef7796d30aff42be9183e7177165b40f90",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libusb1-1.0.29-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libusb1-0__1.0.29-3.el10.s390x",
    sha256 = "9071987b92f299616adeffc5e6b6e0fe4ec0cd1d7cd293cedeb5e06f17ee6a9a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libusb1-1.0.29-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "libusb1-0__1.0.29-3.el10.x86_64",
    sha256 = "60b40b436504fe0b046ea990512936c916c7500f08aaf82e8ea886cb06ca5f53",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libusb1-1.0.29-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libutempter-0__1.2.1-15.el10.aarch64",
    sha256 = "6444bf715fdd137bd1bd096d9903e29516c609d41113139756df2e9316825d6a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libutempter-1.2.1-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libutempter-0__1.2.1-15.el10.s390x",
    sha256 = "1314f6b74597ad5a5a85b51f5243f3802d2f722a5ed5a41a3ebca827eb7d6a6f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libutempter-1.2.1-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "libutempter-0__1.2.1-15.el10.x86_64",
    sha256 = "db498c4b6ce6f223597f8ea955fe4e286f4fc5838e81579de877ca0e80c2d6eb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libutempter-1.2.1-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libuuid-0__2.40.2-15.el10.aarch64",
    sha256 = "f165963e4e47c6a79848df3a775c10607083ca1664f86e3c554e186c7c4b5d3b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libuuid-2.40.2-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libuuid-0__2.40.2-15.el10.s390x",
    sha256 = "37093b4261f004e73d0edd0e2cf36fe68bf947bd6984a8c2a73508ab92169749",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libuuid-2.40.2-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "libuuid-0__2.40.2-15.el10.x86_64",
    sha256 = "df0b13144e4bc5c7b7607d66ac067d411d2f6ef362c9541d20bbda08ba65e371",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libuuid-2.40.2-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libverto-0__0.3.2-10.el10.aarch64",
    sha256 = "0583db7823a8f33a1e09db1e4aa389c10bc98b58de3bd985b6f67be5351d814a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libverto-0.3.2-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libverto-0__0.3.2-10.el10.s390x",
    sha256 = "80757eae2999d4dbc8975747eb4d8fdfb64b144826ba58215672a0f34d313228",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libverto-0.3.2-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "libverto-0__0.3.2-10.el10.x86_64",
    sha256 = "52777e532dc2351c83b72b5033c40df20494afb6504100f7413a65f74368c284",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libverto-0.3.2-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libvirt-client-0__11.10.0-2.el10.aarch64",
    sha256 = "df833b65e4ee6047544d26fb099c8dd977f1c5037f18b26769185c98b6a94870",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libvirt-client-11.10.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libvirt-client-0__11.10.0-2.el10.s390x",
    sha256 = "b078aac179d3c9a77109bb0c517bbc951098951e7ba2e2f429ff6c3abbd958d5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libvirt-client-11.10.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libvirt-client-0__11.10.0-2.el10.x86_64",
    sha256 = "22d5b09ad651276b13a01d6a70527f7e01fe38ea4aacc4e2c4ecaef56263b638",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libvirt-client-11.10.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-common-0__11.10.0-2.el10.aarch64",
    sha256 = "bf4cb56a6935f66e96ff7db2e69c1caadeeda686e36cc1e94101558d5efbb885",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libvirt-daemon-common-11.10.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-common-0__11.10.0-2.el10.s390x",
    sha256 = "b5f611d5be71e57fbcefc6cbd93c0975f8f34b5b4399791a810e7007548057df",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libvirt-daemon-common-11.10.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-common-0__11.10.0-2.el10.x86_64",
    sha256 = "7692c66b827c2115742678e945a6e2439b4172fe36928f8d342af8840d7ed6bc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libvirt-daemon-common-11.10.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-driver-qemu-0__11.10.0-2.el10.aarch64",
    sha256 = "1f7475430f454794794889d5d5e827c07ec842719a74844ea3e797eaaf30e6a8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libvirt-daemon-driver-qemu-11.10.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-driver-qemu-0__11.10.0-2.el10.s390x",
    sha256 = "ed3b8d512b2156297ce05fa5d974b77e2304ce36db5acc9f45e3aeced83f49a5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libvirt-daemon-driver-qemu-11.10.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-driver-qemu-0__11.10.0-2.el10.x86_64",
    sha256 = "77b0d98d943bcf6d09637e8b1888b17b09e451f31a56866d722052e6bd2aa558",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libvirt-daemon-driver-qemu-11.10.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-driver-secret-0__11.10.0-2.el10.s390x",
    sha256 = "2e9b6c638e06624e6e78e316121f91e530d1d5dbf34a27b62fe71758bb42ea3e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libvirt-daemon-driver-secret-11.10.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-driver-secret-0__11.10.0-2.el10.x86_64",
    sha256 = "be3b620e72f39b4782184492396675f0ed95de3d6d82632bb432e127c0a8106f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libvirt-daemon-driver-secret-11.10.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-driver-storage-core-0__11.10.0-2.el10.s390x",
    sha256 = "583dfcaa660df1ae76d49215b2c3b3983e6a8015ba156694b27c81c85a7c87da",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libvirt-daemon-driver-storage-core-11.10.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-driver-storage-core-0__11.10.0-2.el10.x86_64",
    sha256 = "2fe6c3f8c80b180f59cf4d12b5bd94af01506a90f9667dda480698c1aa0a5cdf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libvirt-daemon-driver-storage-core-11.10.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-log-0__11.10.0-2.el10.aarch64",
    sha256 = "4862db80ab75e5306a3443a385926e1f5078f464e31ff36bb8230848e043a441",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libvirt-daemon-log-11.10.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-log-0__11.10.0-2.el10.s390x",
    sha256 = "3da1b34dfa7726533953bf97ddf7f5a1d055c1497bbe746147ee1666594646fd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libvirt-daemon-log-11.10.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libvirt-daemon-log-0__11.10.0-2.el10.x86_64",
    sha256 = "847390056c804637424c0c6d8a62a6e9b4b5ded8369e299b6af6a5a3156f098b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libvirt-daemon-log-11.10.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libvirt-devel-0__11.10.0-2.el10.aarch64",
    sha256 = "1d8309c2ee9c50bfb28602f1e03861348f88cb9258474fe7f968f00e9b717854",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/aarch64/os/Packages/libvirt-devel-11.10.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libvirt-devel-0__11.10.0-2.el10.s390x",
    sha256 = "b04793934320541758e5d8322e0d15a434dbc3c31f45d81ec8103693011ebbae",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/s390x/os/Packages/libvirt-devel-11.10.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libvirt-devel-0__11.10.0-2.el10.x86_64",
    sha256 = "b0738e63829bef9e10c4e76cc2b041aa467026a31f3ac3e52f9d0ee340c8e878",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/x86_64/os/Packages/libvirt-devel-11.10.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libvirt-libs-0__11.10.0-2.el10.aarch64",
    sha256 = "f6d4042b82116a37c4cf8af8611cef3dc91345b41b20e570391890986c851156",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libvirt-libs-11.10.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libvirt-libs-0__11.10.0-2.el10.s390x",
    sha256 = "5c53fff93f7ac5bf69ca0c94ac9e910f0b35bd4d4d2ca2c8fc63a203b9404c6f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libvirt-libs-11.10.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "libvirt-libs-0__11.10.0-2.el10.x86_64",
    sha256 = "9cc0acc8b4c1d3a7bac6e11e6bad344097f1223760cf4c8306a2e4bfd178504a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libvirt-libs-11.10.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libxcrypt-0__4.4.36-10.el10.aarch64",
    sha256 = "465ade16c8f369b5abc1a39671f882bc645ac90b1aeaa29cdfc3958e57640144",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libxcrypt-4.4.36-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libxcrypt-0__4.4.36-10.el10.s390x",
    sha256 = "d14c5523dd6c7f233277acbbb11fb2644f26e91da18e6184ae6ad445e3835a36",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libxcrypt-4.4.36-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "libxcrypt-0__4.4.36-10.el10.x86_64",
    sha256 = "503a29c4c767637d810c7e89ed4355fe0b588381cb360517585fb56a2cf5ee46",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libxcrypt-4.4.36-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libxcrypt-devel-0__4.4.36-10.el10.aarch64",
    sha256 = "2f86c95726f3c3efdcb2d97f5d0020e86d254defebb084df7c13a5fa51442b5a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/libxcrypt-devel-4.4.36-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libxcrypt-devel-0__4.4.36-10.el10.s390x",
    sha256 = "a3f57faa74cefedf8baddec91311a5e0cafe73878e83c3447335495c7ed7934b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libxcrypt-devel-4.4.36-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "libxcrypt-devel-0__4.4.36-10.el10.x86_64",
    sha256 = "ccee1b09985e24bfed47cf7b5c965d7e0e869862ac55f3b7f783cdcec93716f3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libxcrypt-devel-4.4.36-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libxcrypt-static-0__4.4.36-10.el10.aarch64",
    sha256 = "a2d13d4bb5d7ca66384346f0b90801862e9e58e016870565929a699bb1c15feb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/aarch64/os/Packages/libxcrypt-static-4.4.36-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libxcrypt-static-0__4.4.36-10.el10.s390x",
    sha256 = "42d6494724ad9eb96949ac0632b9658b488c5db84d7aa2f9db3d20198a29f9fe",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/s390x/os/Packages/libxcrypt-static-4.4.36-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "libxcrypt-static-0__4.4.36-10.el10.x86_64",
    sha256 = "a4b6f28908fa252bac1c366f91bb37117c0b56ebce1743933dcf2029f10f86c0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/CRB/x86_64/os/Packages/libxcrypt-static-4.4.36-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libxml2-0__2.12.5-9.el10.aarch64",
    sha256 = "4cfe820e11a4e226235d10f4db732bda5c3c44a62e7162766e70819c0ad2aa7d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libxml2-2.12.5-9.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libxml2-0__2.12.5-9.el10.s390x",
    sha256 = "ecfe86d6e6523469d104f42480eddad00802fefeb6bb758f98cd60a9e3c63472",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libxml2-2.12.5-9.el10.s390x.rpm",
    ],
)

rpm(
    name = "libxml2-0__2.12.5-9.el10.x86_64",
    sha256 = "5ae5a9056e911b43ef89472929139ac06fb5fe4a88b03614a221faf6145640e3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libxml2-2.12.5-9.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libxslt-0__1.1.39-8.el10.s390x",
    sha256 = "1e6ec8eb0dac9858a45d3b42ac6755ce77182581ca77af4eec34af9256aa874e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/libxslt-1.1.39-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "libxslt-0__1.1.39-8.el10.x86_64",
    sha256 = "394d4f76d3a0ed6283ecc2e840520958361f9764402d616014687f92ad750d81",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/libxslt-1.1.39-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "libzstd-0__1.5.5-9.el10.aarch64",
    sha256 = "474a4497b7901176be4a59895cd02bba744300fd673668ef068bd1dfc5e129c7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/libzstd-1.5.5-9.el10.aarch64.rpm",
    ],
)

rpm(
    name = "libzstd-0__1.5.5-9.el10.s390x",
    sha256 = "59d29a77a5792bbc4ce42b3ac700a1df776ace058e040f391374f011d39f0eef",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/libzstd-1.5.5-9.el10.s390x.rpm",
    ],
)

rpm(
    name = "libzstd-0__1.5.5-9.el10.x86_64",
    sha256 = "86f3cb406d56283119c45ec8c1f4689aa37ff6c04cf44f6608c10cfdcccdb2c1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/libzstd-1.5.5-9.el10.x86_64.rpm",
    ],
)

rpm(
    name = "lua-libs-0__5.4.6-7.el10.aarch64",
    sha256 = "f8e353910af43a3d81e92ed6355e7d85b64e6946c7af48ac3900bd107e9d91cc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/lua-libs-5.4.6-7.el10.aarch64.rpm",
    ],
)

rpm(
    name = "lua-libs-0__5.4.6-7.el10.s390x",
    sha256 = "e0676e298166577f2766305025f4c99fe774473f084fe13fcdf8937b4b0e5eab",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/lua-libs-5.4.6-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "lua-libs-0__5.4.6-7.el10.x86_64",
    sha256 = "cb9268a17c06928ffb0805ff43d733b0c67171ff3a969cc16b40c7f3e59d64f3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/lua-libs-5.4.6-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "lz4-libs-0__1.9.4-8.el10.aarch64",
    sha256 = "7db176282f02ed0243d66b9136e1269e4db85da61157392ecc0febeac418ec85",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/lz4-libs-1.9.4-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "lz4-libs-0__1.9.4-8.el10.s390x",
    sha256 = "bd0ba485141caa931c930540a150a55a89ab3dfc6bba448aa592e5b9551dee2e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/lz4-libs-1.9.4-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "lz4-libs-0__1.9.4-8.el10.x86_64",
    sha256 = "de360e857e8465c4b38990375e9435efc78e20d022afe42dbf2986d11fc2c759",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/lz4-libs-1.9.4-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "lzo-0__2.10-14.el10.aarch64",
    sha256 = "677b7730dfa8e554a8ddd22940c5c6288b0d51cb09e9547c150905e856fb0575",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/lzo-2.10-14.el10.aarch64.rpm",
    ],
)

rpm(
    name = "lzo-0__2.10-14.el10.s390x",
    sha256 = "32bde43a3a00f4b5d078b2c831270f8cc195664e0e3ab5b1c5bcc6dc802e33d5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/lzo-2.10-14.el10.s390x.rpm",
    ],
)

rpm(
    name = "lzo-0__2.10-14.el10.x86_64",
    sha256 = "9e4f4e6dc19d15eb865805a43f5834b0ce3a405dcc6df0fba72f0b73f59685a2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/lzo-2.10-14.el10.x86_64.rpm",
    ],
)

rpm(
    name = "lzop-0__1.04-16.el10.aarch64",
    sha256 = "e463088918132202d22ada263686b6b723af02b6a49066fd6f9d48cf191cb25e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/lzop-1.04-16.el10.aarch64.rpm",
    ],
)

rpm(
    name = "lzop-0__1.04-16.el10.s390x",
    sha256 = "5eeeda50a19223224ac6de6428853904e6210b0c11223e71aa39848e613bcb0b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/lzop-1.04-16.el10.s390x.rpm",
    ],
)

rpm(
    name = "lzop-0__1.04-16.el10.x86_64",
    sha256 = "925d4dfbf179f00032be3a3a1ec7cf8ed8f9b9b2cd8ea87c2a4da1e97fcfd180",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/lzop-1.04-16.el10.x86_64.rpm",
    ],
)

rpm(
    name = "make-1__4.4.1-9.el10.aarch64",
    sha256 = "4cd069f5132c87ad16d02ff648b6389e3e303b41661362252134519993afc45c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/make-4.4.1-9.el10.aarch64.rpm",
    ],
)

rpm(
    name = "make-1__4.4.1-9.el10.s390x",
    sha256 = "aa138cd7a41f8b054dbecd74462e796b47d58cd9058ad8b56734c0cf242dcd80",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/make-4.4.1-9.el10.s390x.rpm",
    ],
)

rpm(
    name = "make-1__4.4.1-9.el10.x86_64",
    sha256 = "7d0b52fe16c826f8b08656abd70509987e69e2b8a9f0c42fda803d41a9e7c74e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/make-4.4.1-9.el10.x86_64.rpm",
    ],
)

rpm(
    name = "mpdecimal-0__2.5.1-12.el10.aarch64",
    sha256 = "f7755f98208b3f400c950ba46acf568f113029893fede5770d19eedadfa0b3ea",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/mpdecimal-2.5.1-12.el10.aarch64.rpm",
    ],
)

rpm(
    name = "mpdecimal-0__2.5.1-12.el10.s390x",
    sha256 = "2dd0dbab48a3481fab6bcb4554b0854cb66c8d142ef28b8e97b7dfc96d4c2c93",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/mpdecimal-2.5.1-12.el10.s390x.rpm",
    ],
)

rpm(
    name = "mpdecimal-0__2.5.1-12.el10.x86_64",
    sha256 = "7d1762e4770170efa93ff4f7e07cf523f62d3e3378f50d87d7b307cd8a73ee77",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/mpdecimal-2.5.1-12.el10.x86_64.rpm",
    ],
)

rpm(
    name = "mpfr-0__4.2.1-6.el10.aarch64",
    sha256 = "ff42c0656eb7659b733cf29dcec9db96216ce1725bb4f633df360fde860a5b47",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/mpfr-4.2.1-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "mpfr-0__4.2.1-6.el10.s390x",
    sha256 = "0a428bc21172ed27583705623cef0fca8f0ff487ca251b268cb6f0cd8ef95cb9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/mpfr-4.2.1-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "mpfr-0__4.2.1-6.el10.x86_64",
    sha256 = "1285b14028ab77959841b214f3d36800df49c35ce5922f1bb44fe72e34da74f4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/mpfr-4.2.1-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "ncurses-base-0__6.4-14.20240127.el10.aarch64",
    sha256 = "6e439dd9afd65b489675c37f03bdcd950353ed6b822c31ee620fe370642db042",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/ncurses-base-6.4-14.20240127.el10.noarch.rpm",
    ],
)

rpm(
    name = "ncurses-base-0__6.4-14.20240127.el10.s390x",
    sha256 = "6e439dd9afd65b489675c37f03bdcd950353ed6b822c31ee620fe370642db042",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/ncurses-base-6.4-14.20240127.el10.noarch.rpm",
    ],
)

rpm(
    name = "ncurses-base-0__6.4-14.20240127.el10.x86_64",
    sha256 = "6e439dd9afd65b489675c37f03bdcd950353ed6b822c31ee620fe370642db042",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/ncurses-base-6.4-14.20240127.el10.noarch.rpm",
    ],
)

rpm(
    name = "ncurses-libs-0__6.4-14.20240127.el10.aarch64",
    sha256 = "d781030401acc90746bf50c039bac36bcb812b33bb76965d6e6cfed43787a45b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/ncurses-libs-6.4-14.20240127.el10.aarch64.rpm",
    ],
)

rpm(
    name = "ncurses-libs-0__6.4-14.20240127.el10.s390x",
    sha256 = "9dccdd3dc565eb6a75ff4e9d359a7de762e11bf834ebfedf03a815188da8d429",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/ncurses-libs-6.4-14.20240127.el10.s390x.rpm",
    ],
)

rpm(
    name = "ncurses-libs-0__6.4-14.20240127.el10.x86_64",
    sha256 = "80256770f1fb9639ea2e1cc744ba6cdbc6b65850d74c6e66a64fc9bcbb4837f4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/ncurses-libs-6.4-14.20240127.el10.x86_64.rpm",
    ],
)

rpm(
    name = "nftables-1__1.1.5-3.el10.aarch64",
    sha256 = "be675a4acfbf9cf768d95dc25d5390eda069cf2ee8ac774bb81e701cd5ae3135",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/nftables-1.1.5-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "nftables-1__1.1.5-3.el10.s390x",
    sha256 = "ce456b58b2f65c45c8ecaf147eb6936e106a9bc2169e9ba39f5ab760479e42a0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/nftables-1.1.5-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "nftables-1__1.1.5-3.el10.x86_64",
    sha256 = "b04022e2f5e38f6600bf9d3f8cad167704e19bd1191cdbf0dd62db3f15b16a1c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/nftables-1.1.5-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "nmap-ncat-4__7.92-5.el10.aarch64",
    sha256 = "7975edb3d4e9c583a41707bdd6f4d21dee67e571f7b7338352d75cf09130612e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/nmap-ncat-7.92-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "nmap-ncat-4__7.92-5.el10.s390x",
    sha256 = "a0dd5969f49cf59a2448a75498a25007ea270fa25eed6b9d70c1148b88e37a96",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/nmap-ncat-7.92-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "nmap-ncat-4__7.92-5.el10.x86_64",
    sha256 = "30fcce0936e6fad42a6cca2d6999758fd637ad6d7057bd5c155eb8f49af53157",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/nmap-ncat-7.92-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "npth-0__1.6-21.el10.s390x",
    sha256 = "47f1f79ad844c4d845591871bc752bf8677fb257fa2cc4d58778fab215965bf1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/npth-1.6-21.el10.s390x.rpm",
    ],
)

rpm(
    name = "npth-0__1.6-21.el10.x86_64",
    sha256 = "9d5de697dd346d3eeac85008ab93fbfce90ea49342418402959eda90829578d0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/npth-1.6-21.el10.x86_64.rpm",
    ],
)

rpm(
    name = "numactl-libs-0__2.0.19-3.el10.aarch64",
    sha256 = "81016ab56c83cb8c221216794571ae58bb914e21dd3794c242f7ce8a8d8fbf8f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/numactl-libs-2.0.19-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "numactl-libs-0__2.0.19-3.el10.s390x",
    sha256 = "184bd0085cc03d74e317a2aad472fa7638fffc35e4e1a314700e31b00398a6cc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/numactl-libs-2.0.19-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "numactl-libs-0__2.0.19-3.el10.x86_64",
    sha256 = "263ee2cba1d57996778f70045fbc4657067f73edafd6c6b04f4599c3eb12fbfd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/numactl-libs-2.0.19-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "numad-0__0.5-50.20251104git.el10.aarch64",
    sha256 = "942f7db59b047cc56e6c53c5bb9a2a84ba4715088021e85526e973d9485bc8fa",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/numad-0.5-50.20251104git.el10.aarch64.rpm",
    ],
)

rpm(
    name = "numad-0__0.5-50.20251104git.el10.x86_64",
    sha256 = "91def9a46ee7b6ee35a11276987c7eace5b18e97d6d967fe09139e0a01cb0731",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/numad-0.5-50.20251104git.el10.x86_64.rpm",
    ],
)

rpm(
    name = "openldap-0__2.6.10-1.el10.s390x",
    sha256 = "8ccbbb3c19df87e02214012a0fb7eed53455db552b6548e482734f65039b6057",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/openldap-2.6.10-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "openldap-0__2.6.10-1.el10.x86_64",
    sha256 = "c9b225c90d849b679e4ecdc4108703b54749bed23af829d3238f1551dd88fd27",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/openldap-2.6.10-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "openssl-fips-provider-1__3.5.5-1.el10.aarch64",
    sha256 = "8508a817efb181727d4cbed2ef81ddbde1f1da487709f0a6aaaea4ac265acea2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/openssl-fips-provider-3.5.5-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "openssl-fips-provider-1__3.5.5-1.el10.s390x",
    sha256 = "d9844964d1c85617b43d0a9a9cc92a98d50008925abfc6766edecda075085047",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/openssl-fips-provider-3.5.5-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "openssl-fips-provider-1__3.5.5-1.el10.x86_64",
    sha256 = "8cd0358b2f324315431e075aa4f96aba00f2be2fbea15929808d77b38c450b7b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/openssl-fips-provider-3.5.5-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "openssl-libs-1__3.5.5-1.el10.aarch64",
    sha256 = "b57aea518dd32913ada2d53def2aa2b67ddd97f9826c95392a9ff8942c0bc992",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/openssl-libs-3.5.5-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "openssl-libs-1__3.5.5-1.el10.s390x",
    sha256 = "63346962f9adf622da34b1727f33e5abddbedd9ad47dfea2526f2f28ca0dcd47",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/openssl-libs-3.5.5-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "openssl-libs-1__3.5.5-1.el10.x86_64",
    sha256 = "f398762ae421f1aa605af7e3e8770da8b3fa6fbb0afcf74a2d82945dc4670c39",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/openssl-libs-3.5.5-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "osinfo-db-0__20250606-1.el10.s390x",
    sha256 = "44f126f2f67319b5c84345c63150dd5d87ed468cd63b18b4320593505b90b4d1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/osinfo-db-20250606-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "osinfo-db-0__20250606-1.el10.x86_64",
    sha256 = "44f126f2f67319b5c84345c63150dd5d87ed468cd63b18b4320593505b90b4d1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/osinfo-db-20250606-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "osinfo-db-tools-0__1.11.0-8.el10.s390x",
    sha256 = "fed0a8870fb28338db4b8b2bb6f57d44fcbfcaafe88187d787e3bf6cd5f911f0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/osinfo-db-tools-1.11.0-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "osinfo-db-tools-0__1.11.0-8.el10.x86_64",
    sha256 = "31d38586cdd723e3de145e9b03dd1f4eaa2e63681323a7d0bcce3a47cc2e1d62",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/osinfo-db-tools-1.11.0-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "p11-kit-0__0.26.1-1.el10.aarch64",
    sha256 = "e1aad342a866ae8c7b44333c84a82f1830827fad7d1543a6cfccfe9c95b4e6e6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/p11-kit-0.26.1-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "p11-kit-0__0.26.1-1.el10.s390x",
    sha256 = "6ba3c0dac20f1f19e2adf1d345556042196205bc68f98f34b0ff6bdadad2be00",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/p11-kit-0.26.1-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "p11-kit-0__0.26.1-1.el10.x86_64",
    sha256 = "7cd91bff6dca8f9b5620ab55588549a42f09f83668b6313ce38968965c59dad7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/p11-kit-0.26.1-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "p11-kit-trust-0__0.26.1-1.el10.aarch64",
    sha256 = "73ff16cefdcac061ee9c3c5528935989350240aec659a331ec255cc9b09c9f44",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/p11-kit-trust-0.26.1-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "p11-kit-trust-0__0.26.1-1.el10.s390x",
    sha256 = "3d9823071c8cd27c08591be53cbc1dbfeef7c70bfdb95bf40bd9e591a7e21bea",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/p11-kit-trust-0.26.1-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "p11-kit-trust-0__0.26.1-1.el10.x86_64",
    sha256 = "f11e6e6f177fbb0d28083213edbf5a2dd91a9af897ea3d81d10ca41fae50db37",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/p11-kit-trust-0.26.1-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "pam-0__1.6.1-9.el10.aarch64",
    sha256 = "48762bdea0227ff022ec0740b1241147e76d40898a041628e61d20fe8aea344c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/pam-1.6.1-9.el10.aarch64.rpm",
    ],
)

rpm(
    name = "pam-0__1.6.1-9.el10.s390x",
    sha256 = "565a9b5f35ff92d0f91330974e0da3e7322bb87b398a081cef94258e1db17c76",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/pam-1.6.1-9.el10.s390x.rpm",
    ],
)

rpm(
    name = "pam-0__1.6.1-9.el10.x86_64",
    sha256 = "0709423d4705d5f06c4cd6005d205a1a26fb3c4a9d08bbe197fa103924506157",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/pam-1.6.1-9.el10.x86_64.rpm",
    ],
)

rpm(
    name = "pam-libs-0__1.6.1-9.el10.aarch64",
    sha256 = "6d055fc43ae94b745214a675bc077261395a9c9475a66138912b77314fb064fd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/pam-libs-1.6.1-9.el10.aarch64.rpm",
    ],
)

rpm(
    name = "pam-libs-0__1.6.1-9.el10.s390x",
    sha256 = "8927066863a1128bb08750c8e01a26f57b7393454a604436233b1779d55ae655",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/pam-libs-1.6.1-9.el10.s390x.rpm",
    ],
)

rpm(
    name = "pam-libs-0__1.6.1-9.el10.x86_64",
    sha256 = "4f5115114c1bf4882ce2ed5e7629f07a8d0b6e6a93412e9b3a00ab21d8f49973",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/pam-libs-1.6.1-9.el10.x86_64.rpm",
    ],
)

rpm(
    name = "parted-0__3.6-7.el10.s390x",
    sha256 = "bba170cbf71b85ebc299c466f4a15c14ce80fae6a138ec87014b772bab377f02",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/parted-3.6-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "parted-0__3.6-7.el10.x86_64",
    sha256 = "bb339dd10bc7951376243b5a9ae11e18f3ecd235db576739da006a147c6bf412",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/parted-3.6-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "passt-0__0__caret__20251210.gd04c480-2.el10.aarch64",
    sha256 = "b13beb46385431f9d715217c1712c50bd0c96f25d8658d531495954486c04efb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/passt-0%5E20251210.gd04c480-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "passt-0__0__caret__20251210.gd04c480-2.el10.s390x",
    sha256 = "795615af5477338e3835946b9befca2a259a6c33523b4c9fed07c5ac53742a2d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/passt-0%5E20251210.gd04c480-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "passt-0__0__caret__20251210.gd04c480-2.el10.x86_64",
    sha256 = "d7c62dd6032ecc02fd020f1673d79106190531d3c4ed7c8843abd0cf78c2ebbf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/passt-0%5E20251210.gd04c480-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "pcre2-0__10.44-1.el10.3.aarch64",
    sha256 = "23f2a34aa9bc9c8c6662e93d184e07d7e01d45d0fb1b554fd3ed92c03ba2ae3c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/pcre2-10.44-1.el10.3.aarch64.rpm",
    ],
)

rpm(
    name = "pcre2-0__10.44-1.el10.3.s390x",
    sha256 = "7577ec5ef81b0aa96e340c6d292c4b828e957508962ec1ed68c1f048dff3998e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/pcre2-10.44-1.el10.3.s390x.rpm",
    ],
)

rpm(
    name = "pcre2-0__10.44-1.el10.3.x86_64",
    sha256 = "773781e3aa9994fa8d6105ddc0b3d00fdd735bd589a5d9fe40fe96be6a7d89a7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/pcre2-10.44-1.el10.3.x86_64.rpm",
    ],
)

rpm(
    name = "pcre2-syntax-0__10.44-1.el10.3.aarch64",
    sha256 = "71de87112a846df439b0b3381b35fbba8c6e72109c6a4795c1de96e48bbc5d40",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/pcre2-syntax-10.44-1.el10.3.noarch.rpm",
    ],
)

rpm(
    name = "pcre2-syntax-0__10.44-1.el10.3.s390x",
    sha256 = "71de87112a846df439b0b3381b35fbba8c6e72109c6a4795c1de96e48bbc5d40",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/pcre2-syntax-10.44-1.el10.3.noarch.rpm",
    ],
)

rpm(
    name = "pcre2-syntax-0__10.44-1.el10.3.x86_64",
    sha256 = "71de87112a846df439b0b3381b35fbba8c6e72109c6a4795c1de96e48bbc5d40",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/pcre2-syntax-10.44-1.el10.3.noarch.rpm",
    ],
)

rpm(
    name = "pixman-0__0.43.4-2.el10.aarch64",
    sha256 = "dc2c0f98c210e8209690b1d2a4fffa348b6ad22062461f4b3ebc7d7f6dd0246e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/pixman-0.43.4-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "pixman-0__0.43.4-2.el10.s390x",
    sha256 = "269fcda361ff485d379f8a773e47752758b8c58e288f78196f169149570af637",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/pixman-0.43.4-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "pixman-0__0.43.4-2.el10.x86_64",
    sha256 = "c91d0077a917e843a009c69b63793de4b8d2f9a81414f63e319ac31bbf6a08cb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/pixman-0.43.4-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "pkgconf-0__2.1.0-3.el10.aarch64",
    sha256 = "5bd76130128a85e6275c6e56f7e519532425cd2a5d2db7a795a4d1d15f7d0d57",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/pkgconf-2.1.0-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "pkgconf-0__2.1.0-3.el10.s390x",
    sha256 = "010973bdd551e8489eb97446701fdf3100b8dd0b1ea7efd650412d8869d8181a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/pkgconf-2.1.0-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "pkgconf-0__2.1.0-3.el10.x86_64",
    sha256 = "ced8f494b664667d52245ff94ce6c0b2cad135586a36a9ff7f81281d1533f178",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/pkgconf-2.1.0-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "pkgconf-m4-0__2.1.0-3.el10.aarch64",
    sha256 = "4de2147846658c2849aa28f756e5e906a3012be53e656b4a39ae77076286e828",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/pkgconf-m4-2.1.0-3.el10.noarch.rpm",
    ],
)

rpm(
    name = "pkgconf-m4-0__2.1.0-3.el10.s390x",
    sha256 = "4de2147846658c2849aa28f756e5e906a3012be53e656b4a39ae77076286e828",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/pkgconf-m4-2.1.0-3.el10.noarch.rpm",
    ],
)

rpm(
    name = "pkgconf-m4-0__2.1.0-3.el10.x86_64",
    sha256 = "4de2147846658c2849aa28f756e5e906a3012be53e656b4a39ae77076286e828",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/pkgconf-m4-2.1.0-3.el10.noarch.rpm",
    ],
)

rpm(
    name = "pkgconf-pkg-config-0__2.1.0-3.el10.aarch64",
    sha256 = "3d646b74ccc730b097ceba50c5a054a6017b61e354b0e8731c66b5c266d55e40",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/pkgconf-pkg-config-2.1.0-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "pkgconf-pkg-config-0__2.1.0-3.el10.s390x",
    sha256 = "f895f22efbaa5a4f978600b3df480eabab7bb2eea7f0f8e90897ab4d76ae2102",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/pkgconf-pkg-config-2.1.0-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "pkgconf-pkg-config-0__2.1.0-3.el10.x86_64",
    sha256 = "4f5231ffccc59b5f1c42d85cc0dafea9b6901107660931071cf1e65f99af1e0b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/pkgconf-pkg-config-2.1.0-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "policycoreutils-0__3.9-2.el10.aarch64",
    sha256 = "6efb426864c00bad20c92e898916a9bc4a217934b5dcf5e77e08a1e10bb1b88f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/policycoreutils-3.9-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "policycoreutils-0__3.9-2.el10.s390x",
    sha256 = "ca50814a174dc418970e111a10d1c3524075c0df8647b18121bfbd07e26d829a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/policycoreutils-3.9-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "policycoreutils-0__3.9-2.el10.x86_64",
    sha256 = "0dc8d2b0aeb70502b00f3697c80a71e7d8e03a281fa1e77d720d6a1f71d9a9db",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/policycoreutils-3.9-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "polkit-0__125-4.el10.aarch64",
    sha256 = "c7e294ea2b01e7d3f52f3e88d9520f1c57ed7577a220fc1c25092f05c2c2be09",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/polkit-125-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "polkit-0__125-4.el10.s390x",
    sha256 = "746bed9c6883ac33d60f35f0233289dc8e73137cc2e2028a6308b805385fb1e2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/polkit-125-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "polkit-0__125-4.el10.x86_64",
    sha256 = "e4965cc1a34a64e8b5cc6c8738de6c6c0cf2f08f2dced23f9d88427575c3a386",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/polkit-125-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "polkit-libs-0__125-4.el10.aarch64",
    sha256 = "0b7197e7b1c5c394aeb3254f5678c7ec80053566cb367fd6c2d79a11e9a22f37",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/polkit-libs-125-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "polkit-libs-0__125-4.el10.s390x",
    sha256 = "b8715d596f893869f60408286325995dbeeb570286fd397c3c611d86c30a7014",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/polkit-libs-125-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "polkit-libs-0__125-4.el10.x86_64",
    sha256 = "dea631790902108c8e2932c272cb24a5949c3c329009b96cf2f9c8fa5aaee29a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/polkit-libs-125-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "popt-0__1.19-8.el10.aarch64",
    sha256 = "4c727d11de14d8bf1bc0df2be55c75cb0200a685c2737c740e636dedd3edbb0c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/popt-1.19-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "popt-0__1.19-8.el10.s390x",
    sha256 = "3dca46c310266fc9cce48d39651984d726cf727e12f446b70db78cd6f96e3515",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/popt-1.19-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "popt-0__1.19-8.el10.x86_64",
    sha256 = "bd15d2816600655a5241bc3efe6e1ac386061ba6ff2d05e53c70683db8761e5f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/popt-1.19-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "procps-ng-0__4.0.4-9.el10.aarch64",
    sha256 = "41441a1d9724db5a3327dd3fd0dec9a1f5a19345677f346bd4ea295b1657bfc8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/procps-ng-4.0.4-9.el10.aarch64.rpm",
    ],
)

rpm(
    name = "procps-ng-0__4.0.4-9.el10.s390x",
    sha256 = "76e7f6323ac9f9a769a9b27d39673bf2a4cbbc4cf740d186d3b38e3d3f98238e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/procps-ng-4.0.4-9.el10.s390x.rpm",
    ],
)

rpm(
    name = "procps-ng-0__4.0.4-9.el10.x86_64",
    sha256 = "8c618a494766c8c85fee4acd5fa730722bfab04694f72d01219324ec7adb43fb",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/procps-ng-4.0.4-9.el10.x86_64.rpm",
    ],
)

rpm(
    name = "protobuf-c-0__1.5.0-6.el10.aarch64",
    sha256 = "3e17d0b103c0444852bbb952e39a64bb18b36f797cf5d4fd1bea57d9bc2c4cbe",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/protobuf-c-1.5.0-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "protobuf-c-0__1.5.0-6.el10.s390x",
    sha256 = "70c17f3805e9ecb6eaef0c13ae83b889405c22bdf09eafc74d3f0ba26e0882c0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/protobuf-c-1.5.0-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "protobuf-c-0__1.5.0-6.el10.x86_64",
    sha256 = "a7e792e5ed4f89d5f48eb60453619a6e0ad5cd34469526952c1748f1a99ce3ba",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/protobuf-c-1.5.0-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "psmisc-0__23.6-8.el10.aarch64",
    sha256 = "cca0153b72dfb9c9e2f9f8386514ff9591b6166e0746ff32d5cda0eeb9adbaba",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/psmisc-23.6-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "psmisc-0__23.6-8.el10.s390x",
    sha256 = "3c0f3724b4040c7a6c07f5405850ed172f906daa9a7904cd78c8e52c680d4611",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/psmisc-23.6-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "psmisc-0__23.6-8.el10.x86_64",
    sha256 = "9fea410c82d95565a4cbb178da9557ec9cef3512d573efd4dd940c9f2c4219cf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/psmisc-23.6-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "publicsuffix-list-dafsa-0__20240107-5.el10.s390x",
    sha256 = "440cb6e03187dfd68f62abf1dd751ace84ec8e2179d7de45dde348cf2e7dba11",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/publicsuffix-list-dafsa-20240107-5.el10.noarch.rpm",
    ],
)

rpm(
    name = "publicsuffix-list-dafsa-0__20240107-5.el10.x86_64",
    sha256 = "440cb6e03187dfd68f62abf1dd751ace84ec8e2179d7de45dde348cf2e7dba11",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/publicsuffix-list-dafsa-20240107-5.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-0__3.12.12-3.el10.aarch64",
    sha256 = "21775dabf6661090390f8c62c21d436de108d0487e1cc2530c0523b869bd2b45",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-3.12.12-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "python3-0__3.12.12-3.el10.s390x",
    sha256 = "a6bd9d3a578d948bb4deca00b9f0dbc9c6f3d44a318a97ebd469a8d89ca3d39e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-3.12.12-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "python3-0__3.12.12-3.el10.x86_64",
    sha256 = "cd420c83445309ee7ded2740579742da574ab210ed2209d11472a34307fc1b5c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-3.12.12-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "python3-configshell-1__1.1.30-9.el10.aarch64",
    sha256 = "bd3efd00e70a1cfcc68c0d973a5fb3fb34bd9863f30a1330070ba9b718acdf1b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-configshell-1.1.30-9.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-configshell-1__1.1.30-9.el10.s390x",
    sha256 = "bd3efd00e70a1cfcc68c0d973a5fb3fb34bd9863f30a1330070ba9b718acdf1b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-configshell-1.1.30-9.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-configshell-1__1.1.30-9.el10.x86_64",
    sha256 = "bd3efd00e70a1cfcc68c0d973a5fb3fb34bd9863f30a1330070ba9b718acdf1b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-configshell-1.1.30-9.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-dbus-0__1.3.2-8.el10.aarch64",
    sha256 = "6508de73c7fb8966b0d05f631af1002cb5238237791b0bd1b085384b9d6e15fd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-dbus-1.3.2-8.el10.aarch64.rpm",
    ],
)

rpm(
    name = "python3-dbus-0__1.3.2-8.el10.s390x",
    sha256 = "98bf40e2dadd95cc0640650673b54d2dc7ebfa48bbba64a7857aede9c42550a5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-dbus-1.3.2-8.el10.s390x.rpm",
    ],
)

rpm(
    name = "python3-dbus-0__1.3.2-8.el10.x86_64",
    sha256 = "95455a0bc5c76704ba2e46f5dd68b8bd47027b83ae551d62f02b15415a78a164",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-dbus-1.3.2-8.el10.x86_64.rpm",
    ],
)

rpm(
    name = "python3-gobject-base-0__3.46.0-7.el10.aarch64",
    sha256 = "66267f4d40ef4d29b7084c60752688856a78949de1347292d4fe25be501e024b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-gobject-base-3.46.0-7.el10.aarch64.rpm",
    ],
)

rpm(
    name = "python3-gobject-base-0__3.46.0-7.el10.s390x",
    sha256 = "19a92f5cebbd47d89e69c63172d504154790e0ef013967d872ceb7ab0bc4b3f3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-gobject-base-3.46.0-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "python3-gobject-base-0__3.46.0-7.el10.x86_64",
    sha256 = "dd8582c736f50481252e556960f885c63af2d6a64888027d9345e35ec9bc0e27",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-gobject-base-3.46.0-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "python3-gobject-base-noarch-0__3.46.0-7.el10.aarch64",
    sha256 = "2c6d337336442bb1286b43facf50f9d7dad1398f86b312c93a3268d18d230824",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-gobject-base-noarch-3.46.0-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-gobject-base-noarch-0__3.46.0-7.el10.s390x",
    sha256 = "2c6d337336442bb1286b43facf50f9d7dad1398f86b312c93a3268d18d230824",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-gobject-base-noarch-3.46.0-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-gobject-base-noarch-0__3.46.0-7.el10.x86_64",
    sha256 = "2c6d337336442bb1286b43facf50f9d7dad1398f86b312c93a3268d18d230824",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-gobject-base-noarch-3.46.0-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-kmod-0__0.9.2-6.el10.aarch64",
    sha256 = "75e70ceef21104220cd13f4de2cd23669c5660c23a07d15e471b39fa61418e8e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-kmod-0.9.2-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "python3-kmod-0__0.9.2-6.el10.s390x",
    sha256 = "cc4bf90e94d8a7ac36762f21899fc174fe734aae502af314933685c0a342e832",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-kmod-0.9.2-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "python3-kmod-0__0.9.2-6.el10.x86_64",
    sha256 = "c9f50b595ee5a45bb14d571974662ed42c84bcb8660c3a78da891038847da651",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-kmod-0.9.2-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "python3-libs-0__3.12.12-3.el10.aarch64",
    sha256 = "9f56d7dd4675899a19162cce8b778a232afc8f4513b6c69c2e66dd6a4fe0bf5e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-libs-3.12.12-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "python3-libs-0__3.12.12-3.el10.s390x",
    sha256 = "0103b4031a7fb202d5361241ddfcee0d96e0ec33bc7161ca617a9a019d6f30f8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-libs-3.12.12-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "python3-libs-0__3.12.12-3.el10.x86_64",
    sha256 = "d20b719ab3bd3456197544ea9f8e14d0c566987153480ca264ce1e15056b6da8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-libs-3.12.12-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "python3-pip-wheel-0__23.3.2-7.el10.aarch64",
    sha256 = "19b2ce4f91ed680267712a2d2158e679267f9163db71f57e9db5f5c684ac15d8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-pip-wheel-23.3.2-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-pip-wheel-0__23.3.2-7.el10.s390x",
    sha256 = "19b2ce4f91ed680267712a2d2158e679267f9163db71f57e9db5f5c684ac15d8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-pip-wheel-23.3.2-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-pip-wheel-0__23.3.2-7.el10.x86_64",
    sha256 = "19b2ce4f91ed680267712a2d2158e679267f9163db71f57e9db5f5c684ac15d8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-pip-wheel-23.3.2-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-pyparsing-0__3.1.1-7.el10.aarch64",
    sha256 = "8aef56a037934c4132e83b49893c0082351e96ab2c34cf3e14ee41472bb315e2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-pyparsing-3.1.1-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-pyparsing-0__3.1.1-7.el10.s390x",
    sha256 = "8aef56a037934c4132e83b49893c0082351e96ab2c34cf3e14ee41472bb315e2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-pyparsing-3.1.1-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-pyparsing-0__3.1.1-7.el10.x86_64",
    sha256 = "8aef56a037934c4132e83b49893c0082351e96ab2c34cf3e14ee41472bb315e2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-pyparsing-3.1.1-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-pyudev-0__0.24.1-10.el10.aarch64",
    sha256 = "69e5069331c66c49738f7c558b3b78ec5aab81741af1d810629fb3a878a3f540",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-pyudev-0.24.1-10.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-pyudev-0__0.24.1-10.el10.s390x",
    sha256 = "69e5069331c66c49738f7c558b3b78ec5aab81741af1d810629fb3a878a3f540",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-pyudev-0.24.1-10.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-pyudev-0__0.24.1-10.el10.x86_64",
    sha256 = "69e5069331c66c49738f7c558b3b78ec5aab81741af1d810629fb3a878a3f540",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-pyudev-0.24.1-10.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-rtslib-0__2.1.76-12.el10.aarch64",
    sha256 = "4e91e035feac9802dedfe00460866864100d0c32f9b2c2a3a0c32789e307b63c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/python3-rtslib-2.1.76-12.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-rtslib-0__2.1.76-12.el10.s390x",
    sha256 = "4e91e035feac9802dedfe00460866864100d0c32f9b2c2a3a0c32789e307b63c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/python3-rtslib-2.1.76-12.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-rtslib-0__2.1.76-12.el10.x86_64",
    sha256 = "4e91e035feac9802dedfe00460866864100d0c32f9b2c2a3a0c32789e307b63c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/python3-rtslib-2.1.76-12.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-six-0__1.16.0-16.el10.aarch64",
    sha256 = "587391f25be67ed7389c4623f1260a16b33dfab99b5b7376e9eb72dafbc78403",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-six-1.16.0-16.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-six-0__1.16.0-16.el10.s390x",
    sha256 = "587391f25be67ed7389c4623f1260a16b33dfab99b5b7376e9eb72dafbc78403",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-six-1.16.0-16.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-six-0__1.16.0-16.el10.x86_64",
    sha256 = "587391f25be67ed7389c4623f1260a16b33dfab99b5b7376e9eb72dafbc78403",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-six-1.16.0-16.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-typing-extensions-0__4.9.0-6.el10.aarch64",
    sha256 = "d5e02bc63a658039701accb13f243d421082d21a64267e18fd04954d7d2938a8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-typing-extensions-4.9.0-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-typing-extensions-0__4.9.0-6.el10.s390x",
    sha256 = "d5e02bc63a658039701accb13f243d421082d21a64267e18fd04954d7d2938a8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-typing-extensions-4.9.0-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-typing-extensions-0__4.9.0-6.el10.x86_64",
    sha256 = "d5e02bc63a658039701accb13f243d421082d21a64267e18fd04954d7d2938a8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-typing-extensions-4.9.0-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-urwid-0__2.5.3-4.el10.aarch64",
    sha256 = "b754bb3fe723d716e43404b58f706b506b14123fbfd3cad0fa016da10ce7aaf0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-urwid-2.5.3-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "python3-urwid-0__2.5.3-4.el10.s390x",
    sha256 = "e73588c982971858102bb9a6804391709e72e5dd765034f5e4b1d7bea8c59332",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-urwid-2.5.3-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "python3-urwid-0__2.5.3-4.el10.x86_64",
    sha256 = "8ccc08409b227ee8b2cc6879a3b2f84a0e9cc792b720f3b5b7dd1381109a51cd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-urwid-2.5.3-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "python3-wcwidth-0__0.2.6-6.el10.aarch64",
    sha256 = "0477cede1c6397494f32acfba7e6fba166e6f73811cf8c8e62a30aa7b3ae1af8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/python3-wcwidth-0.2.6-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-wcwidth-0__0.2.6-6.el10.s390x",
    sha256 = "0477cede1c6397494f32acfba7e6fba166e6f73811cf8c8e62a30aa7b3ae1af8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/python3-wcwidth-0.2.6-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "python3-wcwidth-0__0.2.6-6.el10.x86_64",
    sha256 = "0477cede1c6397494f32acfba7e6fba166e6f73811cf8c8e62a30aa7b3ae1af8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/python3-wcwidth-0.2.6-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "qemu-img-18__10.1.0-11.el10.aarch64",
    sha256 = "a569f70d8f5c2e1c0693b7310b13c033a303832e908ed1aca0ade126d51e7be2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/qemu-img-10.1.0-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "qemu-img-18__10.1.0-11.el10.s390x",
    sha256 = "2b1d9a00f8deb15450df7ffeba8b2d3852a9a22b2599971223d197a3deb81b90",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/qemu-img-10.1.0-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "qemu-img-18__10.1.0-11.el10.x86_64",
    sha256 = "49c32346581296f9c9341aef0b7ee7bcc948218e336b88f4a6bb30216613d823",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-img-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-common-18__10.1.0-11.el10.aarch64",
    sha256 = "28f2c6c7a45643e8af4cafe028d2b771c0ff6a8033a5629cd2bdf55f99832ba9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/qemu-kvm-common-10.1.0-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-common-18__10.1.0-11.el10.s390x",
    sha256 = "833d96b773e0d69215ab2b975a37af667c9ad1c1f0faa9ea65c684f2e3a0f6dc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/qemu-kvm-common-10.1.0-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "qemu-kvm-common-18__10.1.0-11.el10.x86_64",
    sha256 = "12c7b1d1ae4ed4d32900b6d445812f686017b26860505f01a72f169a56fb36bf",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-kvm-common-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-core-18__10.1.0-11.el10.aarch64",
    sha256 = "a9c162bc2167ebe69056eea04bf99ce3e8ab5660db5d5fc8b3c360585eac83e1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/qemu-kvm-core-10.1.0-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-core-18__10.1.0-11.el10.s390x",
    sha256 = "c31b2474238c901dd63d5ed453764bae4d616ad3f7ea86a8fea32d0d8110ca2f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/qemu-kvm-core-10.1.0-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "qemu-kvm-core-18__10.1.0-11.el10.x86_64",
    sha256 = "86194391f63a26d0421541da474365563adb2019d07e52f4fcc1c1a96dd734bc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-kvm-core-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-display-virtio-gpu-18__10.1.0-11.el10.aarch64",
    sha256 = "3308a97062d4b394298f304222d51ada4eaeceb880b6d08e65b916ac928ae164",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/qemu-kvm-device-display-virtio-gpu-10.1.0-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-display-virtio-gpu-18__10.1.0-11.el10.s390x",
    sha256 = "727b89e3828fbc7bbfaefad0659e5370ea4673284d2afcf0862bbd4f41e8664f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/qemu-kvm-device-display-virtio-gpu-10.1.0-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-display-virtio-gpu-18__10.1.0-11.el10.x86_64",
    sha256 = "2fefdf73cbb5824bc5364ce0affbad48faf2a9d57b6cca784b0b26700a3b8e9b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-kvm-device-display-virtio-gpu-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-display-virtio-gpu-ccw-18__10.1.0-11.el10.s390x",
    sha256 = "d7e9e48ba36616ec8dbfbcc0fc7b62de04200c7a8935c7e311ffc410bf31f42d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/qemu-kvm-device-display-virtio-gpu-ccw-10.1.0-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-display-virtio-gpu-pci-18__10.1.0-11.el10.aarch64",
    sha256 = "460690efde08449e27d99c7a6ec8ed54fed47e5e146e252e7c3975679cab8cab",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/qemu-kvm-device-display-virtio-gpu-pci-10.1.0-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-display-virtio-gpu-pci-18__10.1.0-11.el10.x86_64",
    sha256 = "bc7ad5145e50a5d53d3dcaca9c0e3febe9ef4995b284017b201e0ab59467d060",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-kvm-device-display-virtio-gpu-pci-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-display-virtio-vga-18__10.1.0-11.el10.x86_64",
    sha256 = "eb39072537c28199bd05552e04af2a0a9ea5861e30c392fbb4d426567f06c6ac",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-kvm-device-display-virtio-vga-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-usb-host-18__10.1.0-11.el10.aarch64",
    sha256 = "7adce6b58164570ffe743bc42e0a40872e8b0a2275401862ecb3c599e8488b9e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/qemu-kvm-device-usb-host-10.1.0-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-usb-host-18__10.1.0-11.el10.s390x",
    sha256 = "d3917c4122fef0a85c1395661b6fd318e76debb5d753f2a0b6782d42e50ea7f9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/qemu-kvm-device-usb-host-10.1.0-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-usb-host-18__10.1.0-11.el10.x86_64",
    sha256 = "fd948b93fd42a6de232bd55c838908e44f2185ea137f87843d5c153c4eda88c3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-kvm-device-usb-host-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-usb-redirect-18__10.1.0-11.el10.aarch64",
    sha256 = "b25636328ec64b0192150761f8d619618c43038b5e56422549f4c2b5f3f0104e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/qemu-kvm-device-usb-redirect-10.1.0-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "qemu-kvm-device-usb-redirect-18__10.1.0-11.el10.x86_64",
    sha256 = "363e8c6ab90bb303261e00ebd847fc590078f123a68cfc6f34e8e0700103a5d1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-kvm-device-usb-redirect-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "qemu-pr-helper-18__10.1.0-11.el10.aarch64",
    sha256 = "a4c5a19a7264828f44413d02d2aa140c10ef9ad4c7e4d0f26759660e05b23992",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/qemu-pr-helper-10.1.0-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "qemu-pr-helper-18__10.1.0-11.el10.x86_64",
    sha256 = "86a0b76c0d861db465d13e53e49b18fcce07593fcc5cd6a126a2af52dd557c9a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/qemu-pr-helper-10.1.0-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "readline-0__8.2-11.el10.aarch64",
    sha256 = "a1f1fe411d40cb802c7a3e3b105faffe05c2376563bec2d59c71ed28778684cd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/readline-8.2-11.el10.aarch64.rpm",
    ],
)

rpm(
    name = "readline-0__8.2-11.el10.s390x",
    sha256 = "fd8cc3c7dd19bf773afb0488b2521d1710dfc1d8337fbff55c9f8d84572a4f9f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/readline-8.2-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "readline-0__8.2-11.el10.x86_64",
    sha256 = "d8e2d7c011d0e5c56b6875919ce036605862db02d59d6983d470bc5757021783",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/readline-8.2-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "rpm-0__4.19.1.1-21.el10.aarch64",
    sha256 = "a77920742b41f7215c8bd21f9df7101ebb685bd00cced62b5d2e6fda8c9ca23f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/rpm-4.19.1.1-21.el10.aarch64.rpm",
    ],
)

rpm(
    name = "rpm-0__4.19.1.1-21.el10.s390x",
    sha256 = "dfec4a45376597db60e956a387810a6b93b1538354f38a2fea04f9289dfde19d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/rpm-4.19.1.1-21.el10.s390x.rpm",
    ],
)

rpm(
    name = "rpm-0__4.19.1.1-21.el10.x86_64",
    sha256 = "ef6bd8b6b43a4704e6f14dddd0900f94d111d5fc35a69de78f93999de26c3ed2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/rpm-4.19.1.1-21.el10.x86_64.rpm",
    ],
)

rpm(
    name = "rpm-libs-0__4.19.1.1-21.el10.aarch64",
    sha256 = "753191e6f82ed6e841c8ae435ec2d0b67b8daf31c0e40843cf14645ce5f718c0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/rpm-libs-4.19.1.1-21.el10.aarch64.rpm",
    ],
)

rpm(
    name = "rpm-libs-0__4.19.1.1-21.el10.s390x",
    sha256 = "65f0e8add7fa5e01c16babd21f66cf6b08f429c21f91e4fa1efe3d72932bb674",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/rpm-libs-4.19.1.1-21.el10.s390x.rpm",
    ],
)

rpm(
    name = "rpm-libs-0__4.19.1.1-21.el10.x86_64",
    sha256 = "130985a89e230b91fa829d790de35042843355ad25d41cfa24808296d87047fd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/rpm-libs-4.19.1.1-21.el10.x86_64.rpm",
    ],
)

rpm(
    name = "rpm-sequoia-0__1.9.0.3-1.el10.aarch64",
    sha256 = "493981060d42eb43b76084339119d2ad32c631a69316ccf9cac06e2c01685c17",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/rpm-sequoia-1.9.0.3-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "rpm-sequoia-0__1.9.0.3-1.el10.s390x",
    sha256 = "263f4ca38f6593d00bea74ab4d2ccd7e5a7a478f45b59a8e85c5b0277c83b953",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/rpm-sequoia-1.9.0.3-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "rpm-sequoia-0__1.9.0.3-1.el10.x86_64",
    sha256 = "1e503067eec855e443b262b21b6cda63867a5362814331b8bff3b4f0a0c46b1d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/rpm-sequoia-1.9.0.3-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "scrub-0__2.6.1-11.el10.s390x",
    sha256 = "527e3cf6d20579cbc13efd1b13c639c243e538b6e6121a22efd46ec13d2fa557",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/scrub-2.6.1-11.el10.s390x.rpm",
    ],
)

rpm(
    name = "scrub-0__2.6.1-11.el10.x86_64",
    sha256 = "935258a3ef8ada2d8cba193df349c9d1dc62d38018b9613aab6f727f93655f85",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/scrub-2.6.1-11.el10.x86_64.rpm",
    ],
)

rpm(
    name = "seabios-0__1.17.0-1.el10.x86_64",
    sha256 = "18044b16fa0f0256167f42ba6ab1f8b5ac338747e150d3c9aead064cd28255c9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/seabios-1.17.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "seabios-bin-0__1.17.0-1.el10.x86_64",
    sha256 = "5edf7ad5039c74faab0fe3bc7f9741db6153c6f9ebe3367d20701d4e659d930d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/seabios-bin-1.17.0-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "seavgabios-bin-0__1.17.0-1.el10.x86_64",
    sha256 = "5ed6563e3d13189aa28fe86d0fef8540d61539aa44dd2d5558ca068e79df4ea2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/seavgabios-bin-1.17.0-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "sed-0__4.9-3.el10.aarch64",
    sha256 = "ffa5a588c4b731f4d0f53095e1c26f8aed9cc7c1e40538908b8429dde3405597",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/sed-4.9-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "sed-0__4.9-3.el10.s390x",
    sha256 = "9d736fb53a44b453a669da46a0f99150cfd207b12d466b11891c5838d317abca",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/sed-4.9-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "sed-0__4.9-3.el10.x86_64",
    sha256 = "e0f382e42cee7264161ae86a9f063aed753b2a64cfea78d61ba8c35b6a980995",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/sed-4.9-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "selinux-policy-0__42.1.15-1.el10.aarch64",
    sha256 = "b1f6f0846abc65c7ef73c7b86138a09c7e491c0c2225bb12e1628ece8c7c4eda",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/selinux-policy-42.1.15-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "selinux-policy-0__42.1.15-1.el10.s390x",
    sha256 = "b1f6f0846abc65c7ef73c7b86138a09c7e491c0c2225bb12e1628ece8c7c4eda",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/selinux-policy-42.1.15-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "selinux-policy-0__42.1.15-1.el10.x86_64",
    sha256 = "b1f6f0846abc65c7ef73c7b86138a09c7e491c0c2225bb12e1628ece8c7c4eda",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/selinux-policy-42.1.15-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "selinux-policy-targeted-0__42.1.15-1.el10.aarch64",
    sha256 = "03963934c85fa312233e75132d7ee664b766cda13f8ea1501d19551afb947598",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/selinux-policy-targeted-42.1.15-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "selinux-policy-targeted-0__42.1.15-1.el10.s390x",
    sha256 = "03963934c85fa312233e75132d7ee664b766cda13f8ea1501d19551afb947598",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/selinux-policy-targeted-42.1.15-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "selinux-policy-targeted-0__42.1.15-1.el10.x86_64",
    sha256 = "03963934c85fa312233e75132d7ee664b766cda13f8ea1501d19551afb947598",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/selinux-policy-targeted-42.1.15-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "setup-0__2.14.5-7.el10.aarch64",
    sha256 = "bd7fb604e635ec8e49abc330cb15e9f30dcc1c6f248495308acd83e41896b29e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/setup-2.14.5-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "setup-0__2.14.5-7.el10.s390x",
    sha256 = "bd7fb604e635ec8e49abc330cb15e9f30dcc1c6f248495308acd83e41896b29e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/setup-2.14.5-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "setup-0__2.14.5-7.el10.x86_64",
    sha256 = "bd7fb604e635ec8e49abc330cb15e9f30dcc1c6f248495308acd83e41896b29e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/setup-2.14.5-7.el10.noarch.rpm",
    ],
)

rpm(
    name = "sevctl-0__0.4.3-3.el10.x86_64",
    sha256 = "790b23bb704c9c42b9478b859f909186751771d9c8e864b5b1eb7a0158e690ca",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/sevctl-0.4.3-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "shadow-utils-2__4.15.0-9.el10.aarch64",
    sha256 = "57e87032b85ab8275629c60274b46ca29d425d2feaf5df9131fe05eb914813ed",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/shadow-utils-4.15.0-9.el10.aarch64.rpm",
    ],
)

rpm(
    name = "shadow-utils-2__4.15.0-9.el10.s390x",
    sha256 = "cadac7aec232378077b4858a673566f425d90ef1a04cc69f1fd5dd259c20571e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/shadow-utils-4.15.0-9.el10.s390x.rpm",
    ],
)

rpm(
    name = "shadow-utils-2__4.15.0-9.el10.x86_64",
    sha256 = "ec463d422dc9f65451543bff9b321093f9c1d37ae85735cf71e79317a7e1faa6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/shadow-utils-4.15.0-9.el10.x86_64.rpm",
    ],
)

rpm(
    name = "snappy-0__1.1.10-7.el10.aarch64",
    sha256 = "cc7bc94dc673d8d6d5b4559036648410e790e2c59e2254bc6acd1578fb5e6781",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/snappy-1.1.10-7.el10.aarch64.rpm",
    ],
)

rpm(
    name = "snappy-0__1.1.10-7.el10.s390x",
    sha256 = "03a4ac68f64e146332224557a251a9d051dada615815dd6eb2f4bb22b73826e0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/snappy-1.1.10-7.el10.s390x.rpm",
    ],
)

rpm(
    name = "snappy-0__1.1.10-7.el10.x86_64",
    sha256 = "952dcfbe66d93bece4a4f3753ce721594acbd2af82cd5ca02bf9028375c136b3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/snappy-1.1.10-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "sqlite-libs-0__3.46.1-5.el10.aarch64",
    sha256 = "217f00d515ac790fd028f0fd70a195a288258d0e1157ce6293ab65d29a965cf1",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/sqlite-libs-3.46.1-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "sqlite-libs-0__3.46.1-5.el10.s390x",
    sha256 = "34d97be1a2df9d53a327cec2ca15887897168b880a19a4b7af2b860ad80b35fe",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/sqlite-libs-3.46.1-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "sqlite-libs-0__3.46.1-5.el10.x86_64",
    sha256 = "fa8bd71adaf88ff1b893731fd5f49c949cf3f618332c9b80390113237699f8e7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/sqlite-libs-3.46.1-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "sssd-client-0__2.12.0-1.el10.aarch64",
    sha256 = "38b0c51913fee7468a2e585c2e9ae811cbcd682ac266f3681da43dcd223d7718",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/sssd-client-2.12.0-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "sssd-client-0__2.12.0-1.el10.s390x",
    sha256 = "a7455c34c936bd13af3360275301ad413fc246e65a969cc32b3fb1b25c7870c0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/sssd-client-2.12.0-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "sssd-client-0__2.12.0-1.el10.x86_64",
    sha256 = "83b0d541eaed2737ab42036d731df94094458b6f0fa39db2e23391b947d59ff0",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/sssd-client-2.12.0-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "swtpm-0__0.9.0-2.el10.aarch64",
    sha256 = "2ab32944b56a5d288754d90a3758d667cdc3703631488a2c2f4ac357880bff0b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/swtpm-0.9.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "swtpm-0__0.9.0-2.el10.s390x",
    sha256 = "4fc33cbd8611b571b7952968cd67999ca3d457f7d331290ed5850f70d292f89d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/swtpm-0.9.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "swtpm-0__0.9.0-2.el10.x86_64",
    sha256 = "2754a70eda7d481964e28e610d493f0c705ae966e75dda10ee901a6cf2ef5919",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/swtpm-0.9.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "swtpm-libs-0__0.9.0-2.el10.aarch64",
    sha256 = "e4233f1d21b64737a8c42fecaa652b2388b897a8748b416cf4bd599f30dd7fe2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/swtpm-libs-0.9.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "swtpm-libs-0__0.9.0-2.el10.s390x",
    sha256 = "2cd497257c5a03b6e579f3ada2bef874350a0fbb0aad78cff5d38e247af20c0f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/swtpm-libs-0.9.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "swtpm-libs-0__0.9.0-2.el10.x86_64",
    sha256 = "57b1c9b2ab6540e9504f32e1aa58331fc98ad9476e47b82907f42ab17ab5288a",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/swtpm-libs-0.9.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "swtpm-tools-0__0.9.0-2.el10.aarch64",
    sha256 = "7da6702303b52d8724152e1235f9e3a1eca4dbb7e2dc2ce51f0b32ac6e04aef9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/swtpm-tools-0.9.0-2.el10.aarch64.rpm",
    ],
)

rpm(
    name = "swtpm-tools-0__0.9.0-2.el10.s390x",
    sha256 = "764ead866ad117155085de09e2d0cf5c483ef9aeb134421ae5cc00892aed146e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/swtpm-tools-0.9.0-2.el10.s390x.rpm",
    ],
)

rpm(
    name = "swtpm-tools-0__0.9.0-2.el10.x86_64",
    sha256 = "4c16e59ac5ef48d0e83d6e0a83ba1838c25c6531fb3ef3564bf98643e1e70503",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/swtpm-tools-0.9.0-2.el10.x86_64.rpm",
    ],
)

rpm(
    name = "systemd-0__257-21.el10.aarch64",
    sha256 = "cda7a62d115e0f2b3cfe2aa6379930073a826e7799b8256fd13feebdee3b16ad",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/systemd-257-21.el10.aarch64.rpm",
    ],
)

rpm(
    name = "systemd-0__257-21.el10.s390x",
    sha256 = "446b6ca90a23a4f2211ca381f58e66878a54bdaeec436452c0863f4738a716ef",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/systemd-257-21.el10.s390x.rpm",
    ],
)

rpm(
    name = "systemd-0__257-21.el10.x86_64",
    sha256 = "a6e8c4fb89da61ab8291f1ae72ae65fbf96d08d54bc86733d4ebd37ccc7952f8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/systemd-257-21.el10.x86_64.rpm",
    ],
)

rpm(
    name = "systemd-container-0__257-21.el10.aarch64",
    sha256 = "903869e3c0890ad2dee569aaac34ac17f9ec4f397a016f94311cfe733280db14",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/systemd-container-257-21.el10.aarch64.rpm",
    ],
)

rpm(
    name = "systemd-container-0__257-21.el10.s390x",
    sha256 = "a6f0304a97d995b86d2d99bf723d10c962650e89e25a33a6e1a38a6b97c353ec",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/systemd-container-257-21.el10.s390x.rpm",
    ],
)

rpm(
    name = "systemd-container-0__257-21.el10.x86_64",
    sha256 = "1f8d942392327786cc13036635573c3f7578e1fbd60e7c7746fb25697058e7ba",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/systemd-container-257-21.el10.x86_64.rpm",
    ],
)

rpm(
    name = "systemd-libs-0__257-21.el10.aarch64",
    sha256 = "dd8aad1f044a279282bbaaabce515c8c3f485093be43b36c49c348ce31944655",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/systemd-libs-257-21.el10.aarch64.rpm",
    ],
)

rpm(
    name = "systemd-libs-0__257-21.el10.s390x",
    sha256 = "abeafe36af17ba108485db15560f8983e6bd9f87355b22678891eef109624c9f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/systemd-libs-257-21.el10.s390x.rpm",
    ],
)

rpm(
    name = "systemd-libs-0__257-21.el10.x86_64",
    sha256 = "250ae420e98f000e61b517b22cfc33d292f9cad17f21fd256412060dd8f88dde",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/systemd-libs-257-21.el10.x86_64.rpm",
    ],
)

rpm(
    name = "systemd-pam-0__257-21.el10.aarch64",
    sha256 = "d7bac21680e6cb8ea830bb51a813185a49d94fb01453efd0eb62042349175b0e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/systemd-pam-257-21.el10.aarch64.rpm",
    ],
)

rpm(
    name = "systemd-pam-0__257-21.el10.s390x",
    sha256 = "82861e6b471fd026c08b12d1a93556cb696247bdc2ada50ab7da75a9401533b6",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/systemd-pam-257-21.el10.s390x.rpm",
    ],
)

rpm(
    name = "systemd-pam-0__257-21.el10.x86_64",
    sha256 = "ce3c5eeb84b99b51f818b6b3084cc8eb1cf5f4252f3d9426ff518ee713b63fdd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/systemd-pam-257-21.el10.x86_64.rpm",
    ],
)

rpm(
    name = "tar-2__1.35-10.el10.aarch64",
    sha256 = "7e8eff09bd7f39b2121c8420e5d91109341321bd63ca17311c36fe5f3b42ecf5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/tar-1.35-10.el10.aarch64.rpm",
    ],
)

rpm(
    name = "tar-2__1.35-10.el10.s390x",
    sha256 = "efb27d3706cbc79b151a5337af23ba2184e5b58f14a2ef0c059b369e4a62d12f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/tar-1.35-10.el10.s390x.rpm",
    ],
)

rpm(
    name = "tar-2__1.35-10.el10.x86_64",
    sha256 = "b3201e691a366ff630dd112a02b6f012d7b796ae87c458e5a9341a02c1fed4a9",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/tar-1.35-10.el10.x86_64.rpm",
    ],
)

rpm(
    name = "target-restore-0__2.1.76-12.el10.aarch64",
    sha256 = "aca595d2a389cf5be70543dbe4b428efdced6358fe31d59db7d2608bedfdbde5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/target-restore-2.1.76-12.el10.noarch.rpm",
    ],
)

rpm(
    name = "target-restore-0__2.1.76-12.el10.s390x",
    sha256 = "aca595d2a389cf5be70543dbe4b428efdced6358fe31d59db7d2608bedfdbde5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/target-restore-2.1.76-12.el10.noarch.rpm",
    ],
)

rpm(
    name = "target-restore-0__2.1.76-12.el10.x86_64",
    sha256 = "aca595d2a389cf5be70543dbe4b428efdced6358fe31d59db7d2608bedfdbde5",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/target-restore-2.1.76-12.el10.noarch.rpm",
    ],
)

rpm(
    name = "targetcli-0__2.1.58-5.el10.aarch64",
    sha256 = "687abcde3940a6867baf0ed5f204e383a731fd3d8023ca0672969b80f7a83422",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/targetcli-2.1.58-5.el10.noarch.rpm",
    ],
)

rpm(
    name = "targetcli-0__2.1.58-5.el10.s390x",
    sha256 = "687abcde3940a6867baf0ed5f204e383a731fd3d8023ca0672969b80f7a83422",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/targetcli-2.1.58-5.el10.noarch.rpm",
    ],
)

rpm(
    name = "targetcli-0__2.1.58-5.el10.x86_64",
    sha256 = "687abcde3940a6867baf0ed5f204e383a731fd3d8023ca0672969b80f7a83422",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/targetcli-2.1.58-5.el10.noarch.rpm",
    ],
)

rpm(
    name = "tpm2-tss-0__4.1.3-5.el10.aarch64",
    sha256 = "ed8e49307084dbe71709e5528003811194fd9e00982222a89d289924e7accaea",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/tpm2-tss-4.1.3-5.el10.aarch64.rpm",
    ],
)

rpm(
    name = "tpm2-tss-0__4.1.3-5.el10.s390x",
    sha256 = "a18c9a2026d523cb1ec472a68212db1b3198a64e8a306548f0341c69bec6cbfa",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/tpm2-tss-4.1.3-5.el10.s390x.rpm",
    ],
)

rpm(
    name = "tpm2-tss-0__4.1.3-5.el10.x86_64",
    sha256 = "863c36c073642a98f267bd3503fb2a505412341f2ae3c828f80e8188b4dc6d97",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/tpm2-tss-4.1.3-5.el10.x86_64.rpm",
    ],
)

rpm(
    name = "tzdata-0__2025c-1.el10.aarch64",
    sha256 = "f42431990a112a5a422eae042de7d28bd2e9d9a971d9082771962b18a2951846",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/tzdata-2025c-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "tzdata-0__2025c-1.el10.s390x",
    sha256 = "f42431990a112a5a422eae042de7d28bd2e9d9a971d9082771962b18a2951846",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/tzdata-2025c-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "tzdata-0__2025c-1.el10.x86_64",
    sha256 = "f42431990a112a5a422eae042de7d28bd2e9d9a971d9082771962b18a2951846",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/tzdata-2025c-1.el10.noarch.rpm",
    ],
)

rpm(
    name = "unbound-libs-0__1.20.0-15.el10.aarch64",
    sha256 = "e3b8abdb07e727487f1a39e6de652c621e87c5279c361a26d852fc89d5430d86",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/unbound-libs-1.20.0-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "unbound-libs-0__1.20.0-15.el10.s390x",
    sha256 = "5dce54a982a4f63e22ac639fc17f027480d81af0916053f4608042b22125bc49",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/unbound-libs-1.20.0-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "unbound-libs-0__1.20.0-15.el10.x86_64",
    sha256 = "95b37ceb9b0a300e1a2894bb18c621db0a9aa12581cefea1eb4086095b0cb7d7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/unbound-libs-1.20.0-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "usbredir-0__0.13.0-6.el10.aarch64",
    sha256 = "12a672104464f85388819600c8b2e4eee38cdb67342107485183bc2076b54fe2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/usbredir-0.13.0-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "usbredir-0__0.13.0-6.el10.x86_64",
    sha256 = "11551f45b3e60a80530431dfcd5a1d29c5624d34aeaf86dcfd6bcd0a4f87337f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/usbredir-0.13.0-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "userspace-rcu-0__0.14.0-7.el10.aarch64",
    sha256 = "4c68e72d9cf6b3ae7b001c181998eeff4514058621e7517ebff26f315757c11d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/userspace-rcu-0.14.0-7.el10.aarch64.rpm",
    ],
)

rpm(
    name = "userspace-rcu-0__0.14.0-7.el10.x86_64",
    sha256 = "2ff9144b446e979b4d014fac8912e7ea9f9dbc2ebbe913c715629bf82aa34082",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/userspace-rcu-0.14.0-7.el10.x86_64.rpm",
    ],
)

rpm(
    name = "util-linux-0__2.40.2-15.el10.aarch64",
    sha256 = "25ba1628a53deba99c20c5149678b950e2ebc60fb8612d2c22d7f6d05686b62f",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/util-linux-2.40.2-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "util-linux-0__2.40.2-15.el10.s390x",
    sha256 = "9c838c66bb698a62102c9ee144069450c3482990ebcbbe5dccff249d8726d1bc",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/util-linux-2.40.2-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "util-linux-0__2.40.2-15.el10.x86_64",
    sha256 = "8e486e9240aaede3947025a502a3f1b43d4aac04ab7400395d923251274ea76c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/util-linux-2.40.2-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "util-linux-core-0__2.40.2-15.el10.aarch64",
    sha256 = "fea3374abdfaa5967ff585a7e8a06dd4612e6dc20ea3909658e04d95547e647e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/util-linux-core-2.40.2-15.el10.aarch64.rpm",
    ],
)

rpm(
    name = "util-linux-core-0__2.40.2-15.el10.s390x",
    sha256 = "7e0133dd073238098b41fd6571da2ad71141c5e58eae5ba682b21b01d099e3a8",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/util-linux-core-2.40.2-15.el10.s390x.rpm",
    ],
)

rpm(
    name = "util-linux-core-0__2.40.2-15.el10.x86_64",
    sha256 = "d5192dee9734c527a4527b9cba6d754fd18b38f1e85d9ca0f2237e647409f3d3",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/util-linux-core-2.40.2-15.el10.x86_64.rpm",
    ],
)

rpm(
    name = "vim-data-2__9.1.083-6.el10.aarch64",
    sha256 = "09b0f20af6272c4f242bce1d67b15c743f625adb78a78ae20425fc877045ae83",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/vim-data-9.1.083-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "vim-data-2__9.1.083-6.el10.s390x",
    sha256 = "09b0f20af6272c4f242bce1d67b15c743f625adb78a78ae20425fc877045ae83",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/vim-data-9.1.083-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "vim-data-2__9.1.083-6.el10.x86_64",
    sha256 = "09b0f20af6272c4f242bce1d67b15c743f625adb78a78ae20425fc877045ae83",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/vim-data-9.1.083-6.el10.noarch.rpm",
    ],
)

rpm(
    name = "vim-minimal-2__9.1.083-6.el10.aarch64",
    sha256 = "b542c48e8187bb909aca7deefc8f1358205744b3faa765df22890c77d5b4467d",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/vim-minimal-9.1.083-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "vim-minimal-2__9.1.083-6.el10.s390x",
    sha256 = "520b48c16edcfe4a1ad88d7481001fc7c03f8c428c4ebd12579a8bb99459ec27",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/vim-minimal-9.1.083-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "vim-minimal-2__9.1.083-6.el10.x86_64",
    sha256 = "9e39250b2c331a51c63c47b319fa85d802d00e8d57e7d68992d28e98ebb09e6c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/vim-minimal-9.1.083-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "virtiofsd-0__1.13.3-1.el10.aarch64",
    sha256 = "ad5eec8ff18d9610a2eff000406908876ea1b1af69cbf97c396b74efc65d54dd",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/virtiofsd-1.13.3-1.el10.aarch64.rpm",
    ],
)

rpm(
    name = "virtiofsd-0__1.13.3-1.el10.s390x",
    sha256 = "a9bd279fd632f35e33ba50593ab833ec2673878ec8e54a45d7e8e20c61fa7ed4",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/virtiofsd-1.13.3-1.el10.s390x.rpm",
    ],
)

rpm(
    name = "virtiofsd-0__1.13.3-1.el10.x86_64",
    sha256 = "fa45976edcd696c9fcda96b0c47b1e91d7c59ae83f5616bd047a0bad6b0be0ae",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/virtiofsd-1.13.3-1.el10.x86_64.rpm",
    ],
)

rpm(
    name = "which-0__2.21-44.el10.aarch64",
    sha256 = "369a215b68f7dd87ce2b0c7be20425b63a19ba8b18a74775b474a717524388fe",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/which-2.21-44.el10.aarch64.rpm",
    ],
)

rpm(
    name = "which-0__2.21-44.el10.s390x",
    sha256 = "93c0edc58db280e4bcc7a7568fc9eb935a27666fea4659c0de6aa93e1017d0d7",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/which-2.21-44.el10.s390x.rpm",
    ],
)

rpm(
    name = "which-0__2.21-44.el10.x86_64",
    sha256 = "8817b5d8ce0a8a07e38daa93a72d0cca53934e1631322b990d389ccb34376e1c",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/which-2.21-44.el10.x86_64.rpm",
    ],
)

rpm(
    name = "xorriso-0__1.5.6-6.el10.aarch64",
    sha256 = "8e152db322abfb8b173f703a0af4be1ef294abeb3dd78da974f800b074a06530",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/aarch64/os/Packages/xorriso-1.5.6-6.el10.aarch64.rpm",
    ],
)

rpm(
    name = "xorriso-0__1.5.6-6.el10.s390x",
    sha256 = "ab7d3d43d22e8a4920453c73de0e60dbc3df69ab28042cad271df8a51cfa0b4b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/s390x/os/Packages/xorriso-1.5.6-6.el10.s390x.rpm",
    ],
)

rpm(
    name = "xorriso-0__1.5.6-6.el10.x86_64",
    sha256 = "2077b91e476836bec242f0fbf4a83384bfc785e9531bedb81ee825936105b017",
    urls = [
        "http://mirror.stream.centos.org/10-stream/AppStream/x86_64/os/Packages/xorriso-1.5.6-6.el10.x86_64.rpm",
    ],
)

rpm(
    name = "xz-1__5.6.2-4.el10.aarch64",
    sha256 = "7bf62608392ae9fd5dd59add39723086f5a052c2064e0498c1641c572cd46460",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/xz-5.6.2-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "xz-1__5.6.2-4.el10.s390x",
    sha256 = "37e1052ce13b55ef1f4e33a8997728963f51c76223d165e2534f0cd6e8f9ba59",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/xz-5.6.2-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "xz-1__5.6.2-4.el10.x86_64",
    sha256 = "dc71c8e5b558c9f9fdea14a7d38819fc12ad8bdcb6834989188b225ca191eded",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/xz-5.6.2-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "xz-libs-1__5.6.2-4.el10.aarch64",
    sha256 = "fcf207b0e6fe443fafe62fa43fc44ce16c8c118dd5e69491b3ad4b9eda72cc61",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/xz-libs-5.6.2-4.el10.aarch64.rpm",
    ],
)

rpm(
    name = "xz-libs-1__5.6.2-4.el10.s390x",
    sha256 = "7edd13c2a8dfb66b1e8c8a0d1d9259a1ff5cfb4891d568cc8f990664a02f7e32",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/xz-libs-5.6.2-4.el10.s390x.rpm",
    ],
)

rpm(
    name = "xz-libs-1__5.6.2-4.el10.x86_64",
    sha256 = "21733e8b6bf26b20633618adb074706972479080527f3f7a51246e83b3d4342e",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/xz-libs-5.6.2-4.el10.x86_64.rpm",
    ],
)

rpm(
    name = "zlib-ng-compat-0__2.2.3-3.el10.aarch64",
    sha256 = "a7870bf73b68086ae1fdd3e2fb6191bf79dff1ab5ae16b907efbb0befe590dca",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/zlib-ng-compat-2.2.3-3.el10.aarch64.rpm",
    ],
)

rpm(
    name = "zlib-ng-compat-0__2.2.3-3.el10.s390x",
    sha256 = "89c8decb9febd474ba2f3fbb38c37577dd7098b349a9e766267723fb94f25962",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/zlib-ng-compat-2.2.3-3.el10.s390x.rpm",
    ],
)

rpm(
    name = "zlib-ng-compat-0__2.2.3-3.el10.x86_64",
    sha256 = "8fe3c2d5203810828fa3e4a5d84ae53172ffd27f4f0eec9d192b42b187795c09",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/zlib-ng-compat-2.2.3-3.el10.x86_64.rpm",
    ],
)

rpm(
    name = "zstd-0__1.5.5-9.el10.aarch64",
    sha256 = "b45bf236f2a5a034295eb933b3c056b302785fa122ecba44b98fec6d2d8b39a2",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/aarch64/os/Packages/zstd-1.5.5-9.el10.aarch64.rpm",
    ],
)

rpm(
    name = "zstd-0__1.5.5-9.el10.s390x",
    sha256 = "def19135b3b6f01e46d9ee17e69ae1227e9addaf1fcd231596836000917fe393",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/s390x/os/Packages/zstd-1.5.5-9.el10.s390x.rpm",
    ],
)

rpm(
    name = "zstd-0__1.5.5-9.el10.x86_64",
    sha256 = "4ef415b98ddbe28f836b86699f4cec6002817ea20fb47499d3c6bb0814db6d4b",
    urls = [
        "http://mirror.stream.centos.org/10-stream/BaseOS/x86_64/os/Packages/zstd-1.5.5-9.el10.x86_64.rpm",
    ],
)
