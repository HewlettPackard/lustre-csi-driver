@Library('dst-shared@master') _

// See https://github.hpe.com/hpe/hpc-dst-jenkins-shared-library for all
// the inputs to the dockerBuildPipeline.
// In particular: vars/dockerBuildPipeline.groovy
dockerBuildPipeline {
        repository = "cray"
        imagePrefix = "cray"
        app = "dp-lustre-csi-driver"
        name = "dp-lustre-csi-driver"
        description = "Operator for lustre filesystem CSI driver"
        dockerfile = "Dockerfile"
        autoJira = false
        createSDPManifest = false
        product = "rabsw"
}
