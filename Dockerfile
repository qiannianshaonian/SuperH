# 打包依赖阶段使用golang作为基础镜像
FROM golang:1.14 as builder

# 启用go module
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app
RUN rm -rf network

COPY . .

# 指定OS等，并go build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

# 由于我不止依赖二进制文件，还依赖views文件夹下的html文件还有assets文件夹下的一些静态文件
# 所以我将这些文件放到了publish文件夹
RUN mkdir publish  && cp SuperH publish
#    && \ cp -r sugar-server.ini publish

#RUN ["chmod", "+x", "/app/publish/entrypoint.sh"] &&   cp -r entrypoint.sh publish
# 运行阶段指定scratch作为基础镜像
FROM alpine

WORKDIR /app

# 将上一个阶段publish文件夹下的所有文件复制进来
COPY --from=builder /app/publish .

# 指定运行时环境变量
# ENV GIN_MODE=release \
#     PORT=80
#
# EXPOSE 80

ENTRYPOINT ["./SuperH"]
