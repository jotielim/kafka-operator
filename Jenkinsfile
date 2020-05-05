// --------------- CHANGE THESE TO FIT YOUR PROJECT ---------------
// The name of your repo
def project_name = "kafka-operator"
// The path in GitLab
def project_dir = "pandora/services/"
// Domain (if necessary)
def project_domain = "gitlab-edge.eng.hitachivantara.com/"
// GitLab SSH
def gitlab_ssh = "git@gitlab-edge.eng.hitachivantara.com:yonathan/kafka-operator.git"
// Name you wish to give your binary
def binary_name = "kafka-operator"
// Image name
def image_name = "kafka-operator"
// Dockerfile name for building your image
def dockerfile_name = "Dockerfile"
// ----------------------------------------------------------------

def image_tag = "$image_tag"
def commit_tag = "$commit_tag"
def build_branch = "$build_branch"
def docker_image = null
def commit_author = null
def commit_message = null
def commit_sha = null
def auto_deploy = "$auto_deploy"
def project_path = "/go/src/$project_domain$project_dir$project_name"
def failure_step = ""

node {

    try {

        // Needed for trigger release builds or general pipeline builds
        if (image_tag == "") {
            image_tag = currentBuild.id
        }

        checkout scm

        // Build docker image to build binary in
        docker.withRegistry('http://lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com', 'artifactory') {
            gitlabCommitStatus(name: 'Prepare Docker Image') {
                stage('Prepare Docker Image') {
                    failure_step = "Prepare Docker Image"
                    docker_image = docker.image("lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/build_images/hiota_image:latest")
                }
            }
        }

        // Work in a docker image that we will run in
        docker_image.inside('-u root -v $WORKSPACE:/output') {

            // Get the Go Stuff
            gitlabCommitStatus(name: 'Checkout Code') {
                stage('Checkout Code') {

                    failure_step = "Checkout Code"

                        sh "mkdir -p $project_path"

                        checkout([$class: 'GitSCM', branches: [[name: build_branch]], doGenerateSubmoduleConfigurations: false, extensions: [[$class: 'RelativeTargetDirectory', relativeTargetDir: "$project_name"]], submoduleCfg: [], userRemoteConfigs: [[credentialsId: '6e5c3081-96ac-4a13-9c02-5af7ea7bcabd', url: "$gitlab_ssh"]]])

                        if (commit_tag != "") {
                            sh "cd $project_name && git reset --hard $commit_tag"
                            echo "Building with SHA $commit_tag"
                        }

                        commit_author = CommitAuthor("$project_name")
                        commit_message = CommitMessage("$project_name")
                        commit_sha = CommitHash("$project_name")

                        sh "mv $project_name/* $project_path"
                        sh "rm -Rf $project_name"
                }
            }

            // Test
            gitlabCommitStatus(name: 'Test') {
                stage('Test') {
                    failure_step = "Test"

                    sh "go version"
                    sh "make"
                    sh "gocov convert cover.out | gocov-xml > kafka-operator-coverage.xml"

                    sh "cp $project_path/test/output/kafka-operator-coverage.xml ."
                    cobertura autoUpdateHealth: false, autoUpdateStability: false, coberturaReportFile: "kafka-operator-coverage.xml", lineCoverageTargets: '70, 0, 70', maxNumberOfBuilds: 0, methodCoverageTargets: '70, 0, 70', onlyStable: false, sourceEncoding: 'ASCII', zoomCoverageChart: false
                }
            }
        }

        // Deploy the artifact to nexus
        gitlabCommitStatus(name: 'Nexus') {
            stage('Nexus') {
                failure_step = "Nexus"

                withCredentials([usernamePassword(credentialsId: 'nexus_login', usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
                    sh "make docker-build IMG=$image_name"

                    sh "docker login lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com --username $USERNAME --password $PASSWORD "

                    sh "docker tag $image_name lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/kafka/$image_name:latest"
                    sh "docker tag $image_name lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/kafka/$image_name:$image_tag"
                    sh "docker tag $image_name lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/kafka/$image_name:$BUILD_NUMBER"

                    sh "docker push lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/kafka/$image_name:latest"
                    sh "docker push lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/kafka/$image_name:$image_tag"
                    sh "docker push lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/kafka/$image_name:$BUILD_NUMBER"

                    sh "docker rmi -f lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/kafka/$image_name:latest"
                }
            }
        }

        // Deploy the Docker containers to the K8s DEV Env.
        gitlabCommitStatus(name: 'Deploy to DEV K8s') {
            stage('Deploy to DEV K8s') {
                failure_step = "Deploy to DEV K8s"

                if ( auto_deploy == "true" ) {
                    String[] hostnames = DEV_BOX_HOSTNAMES.split(",")
                    String[] host_ips = DEV_BOX_IPS.split(",")

                    for(int i = 0; i < hostnames.size(); i++) {

                        println(hostnames[i])
                        println(host_ips[i])

                        withCredentials([usernamePassword(credentialsId: "${hostnames[i]}", passwordVariable: 'PASSWORD', usernameVariable: 'USERNAME')]) {
                            def remote = [:]
                            remote.name = "${hostnames[i]}"
                            remote.host = "${host_ips[i]}"
                            remote.user = "$USERNAME"
                            remote.password = "$PASSWORD"
                            remote.allowAnyHosts = true

                            if ( "${host_ips[i]}" == MULTI_NODE ) {
                                sshCommand remote: remote, command: "docker pull lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/$image_name:$image_tag"
                                sshCommand remote: remote, command: "docker tag lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/$image_name:$image_tag localhost:32500/repository/pandora/$image_name:$image_tag"
                                sshCommand remote: remote, command: "docker push localhost:32500/repository/pandora/$image_name:$image_tag"
                                sshCommand remote: remote, command: "kubectl set image -n hiota deployment/$image_name $image_name=localhost:32500/repository/pandora/$image_name:$image_tag --record=true"
                                sshCommand remote: remote, command: "docker rmi lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/$image_name:$image_tag"
                            } else {
                                sshCommand remote: remote, command: 'hostname'
                                sshCommand remote: remote, command: "docker pull lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/$image_name:$image_tag"
                                sshCommand remote: remote, command: "kubectl set image -n hiota deployment/$image_name $image_name=lumadaedge-docker-dev-sc.repo.sc.eng.hitachivantara.com/repository/pandora/$image_name:$image_tag --record=true"
                            }
                        }
                    }
                }
            }
            currentBuild.result = "SUCCESS"
        }
    }

    catch (Exception e) {
        echo e.message
        currentBuild.result = "FAILURE"
    }

    finally {
        if (currentBuild.result == "FAILURE") {
            slackSend channel: 'pandora_build', message: "${project_name} - Build failed on step ${failure_step}. Changes made by @${commit_author}. Code changes were ${commit_message}. Results at ${env.BUILD_URL} :hankey:", teamDomain: 'hitachivantara-eng', tokenCredentialId: 'JenkinsSlackIntegration'
        }
        else if (currentBuild.result == "UNSTABLE") {
            slackSend channel: 'pandora_build', message: "${project_name} - Build unstable.Changes made by @${commit_author}. Code changes were ${commit_message}.  Results at ${env.BUILD_URL} :face_with_monocle:", teamDomain: 'hitachivantara-eng', tokenCredentialId: 'JenkinsSlackIntegration'
        }
        else if (currentBuild.result == "SUCCESS") {
           // slackSend channel: 'pandora_build', message: "${project_name} - Build succeeded. Results at ${env.BUILD_URL} :parrot:", teamDomain: 'hitachivantara-eng', tokenCredentialId: 'JenkinsSlackIntegration'
        }
        cleanWs()
    }
}

def CommitAuthor(location) {
    sh """cd $location && git --no-pager show -s --format='%ae' >> author"""
    def author = readFile("$location/author").trim()
    sh "rm $location/author"
    author
}

def CommitMessage(location) {
    sh """cd $location && git --no-pager show -s --format='%s' >> message"""
    def message = readFile("$location/message").trim()
    sh "rm $location/message"
    message
}

def CommitHash(location) {
    sh """cd $location && git --no-pager show -s --format='%H' >> commit"""
    def commit = readFile("$location/commit").trim()
    sh "rm $location/commit"
    commit
}
