FROM golang:1.23

WORKDIR /app

ARG TARGETPLATFORM

RUN echo "$TARGETPLATFORM"

# COPY .env ./

COPY Taskfile.yml ./
COPY go.mod ./

COPY internal/ ./internal
COPY cmd/ ./cmd
COPY templates ./templates

RUN go mod download && go mod verify

# RUN mkdir bin

# Install Task as build system
# install script apparently does not work on 32-bit Raspi so this ugly thing is needed
# see https://github.com/go-task/task/issues/1516#issuecomment-2347395883
RUN if [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
    curl -L https://taskfile.dev/install.sh > install-task.sh \
    && chmod +x ./install-task.sh \
    && ./install-task.sh -b $HOME/.local/bin v3.40.0 \
    && $HOME/.local/bin/task all; \
    elif [ "$TARGETPLATFORM" = "linux/arm/v7" ]; then \
    wget https://github.com/go-task/task/releases/download/v3.40.0/task_linux_arm.deb \
    && dpkg -i task_linux_arm.deb \
    && rm task_linux_arm.deb \
    && /usr/bin/task all; \
  fi

EXPOSE 8080

CMD ["bin/golist_svr"]
