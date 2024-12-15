FROM golang:1.23-bookworm

ENV DEBIAN_FRONTEND=noninteractive

ARG GCC_VERSION=11
ARG GPP_VERSION=11

# 更新软件源列表并安装构建工具和 SSH 服务
RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    openssh-server \
    sudo \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# 配置 SSH 服务
RUN mkdir /var/run/sshd && \
    echo 'root:root' | chpasswd && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    echo "UseDNS no" >> /etc/ssh/sshd_config && \
    echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config

# 设置开发用户（非root），密码为developer
RUN useradd -m developer && \
    echo 'developer:developer' | chpasswd && \
    usermod -aG sudo developer

# 推荐：设置 SSH 公钥认证（需要手动配置密钥）
# COPY your_public_key.pub /home/developer/.ssh/authorized_keys

# 暴露18022端口用于SSH连接
EXPOSE 18022

# 设置工作目录
WORKDIR /workspace

CMD ["/usr/sbin/sshd", "-D", "-p", "18022"]
