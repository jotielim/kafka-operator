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
def failure_step = ""
def nexus_ip = "10.76.48.106:5000"

node {

    try {

        // Needed for trigger release builds or general pipeline builds
        if (image_tag == "") {
            image_tag = currentBuild.id
        }

        checkout([$class: 'GitSCM', branches: [[name: build_branch]], doGenerateSubmoduleConfigurations: false, extensions: [], submoduleCfg: [], userRemoteConfigs: [[credentialsId: '6e5c3081-96ac-4a13-9c02-5af7ea7bcabd', url: gitlab_ssh]]])

        // Deploy the artifact to nexus
        gitlabCommitStatus(name: 'Nexus') {
        stage('Nexus') {

            failure_step = "Nexus"

            withCredentials([usernamePassword(credentialsId: 'nexus_login', usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
                sh "make docker-build IMG=$image_name"

                sh "docker login 10.76.48.106:5000 --username $USERNAME --password $PASSWORD "

                sh "docker tag $image_name 10.76.48.106:5000/repository/pandora/kafka/$image_name:0.9.2 "

                sh "docker push 10.76.48.106:5000/repository/pandora/kafka/$image_name:0.9.2"

                sh "docker rmi -f 10.76.48.106:5000/repository/pandora/$image_name:v1"
            }
        }
    }

        // Deploy the Docker containers to the K8s DEV Env.
        gitlabCommitStatus(name: 'Deploy to DEV K8s') {
        stage('Deploy to DEV K8s') {

            failure_step = "Deploy to DEV K8s"
        }
        currentBuild.result = "SUCCESS"
        }
    }

    catch (Exception e) {
        echo e.message
        currentBuild.result = "FAILURE"
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
