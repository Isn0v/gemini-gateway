# Gemini Gateway

Небольшой клиент-серверный проект, где сервер - шлюз между Gemini API и пользователем, а клиент - сам пользователь, который отправляет запросы нейронной сети.
Цель проекта заключается в познании логики работы с таким стеком технологий, как Kubernetes, Helm, Prometheus, Grafana. Другии словами, DevOps инструментами.

## Запуск

Для разворачивания k8s кластера я особо не парился и использовал одноузловой кластер `minikube`:

```bash
minikube config set driver docker
minikube config set memory 6144
minikube config set cpus 4
minikube start
```

## Деплой сервера

Начать стоит с самого главного компонента - сервера. API ключ хранится в секрете, поэтому в деплое используется `Secret`:

```bash
kubectl create secret generic gemini-api-key-secret \
-n default \
--from-literal=api-key=<your-api-key>
```

Далее деплоим сам сервер:
> **Важно:** Помимо самого сервера нужно также настроить `Service` и `Ingress` для доступа к нему извне. В случае `Ingress` нужно добавить имя хоста `hosts`, по которому он будет перенаправлять: 

```bash
echo "gemini-gateway.local $(minikube ip)" | sudo tee -a /etc/hosts

cd $REPO_ROOT/infra/kubernetes/manifests/server

kubectl apply -f gemini-server-deployment.yaml
kubectl apply -f gemini-server-service.yaml
kubectl apply -f gemini-ingress.yaml
```

## Деплой мониторинга за кластером

Здесь я начинаю использовать `Helm`, а конкретно `prometheus` и `grafana`. Для этого есть довольно удобный чарт со всем необходимым, который сам настраивается и понимает что мониторить:

>**Примечание (TODO):** В проецессе я написал отдельный манифест под настройку `Ingress` для `Grafana`, но только под самый конец разработки я узнал, что его можно настроить также в `helm values`.

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

helm repo update

cd $REPO_ROOT/infra/kubernetes/helm/values

helm install prometheus prometheus-community/kube-prometheus-stack
--namespace monitoring  --create-namespace \
-f prometheus-values.yaml 

cd $REPO_ROOT/infra/kubernetes/manifests/monitoring

kubectl apply -f monitoring-ingress.yaml

echo "grafana.my-project.com $(minikube ip)" | sudo tee -a /etc/hosts
```

## Деплой логгера

Здесь я воспользовался `ELK` стеком (Elasticsearch, Logstash, Kibana, Filebeat).

> **Примечание:** Обычно так выходило, что поды не могли корректно перейти в состояние готовности по самым разным причинам. Поэтому чтобы все было круто, нужно выполнять установку чартов в четком порядке, в каком они написаны ниже, убеждаясь, что все поды из предыдущих чартов прошли проверку готовности, и уже только потом приступать к установке следующих. Мониторинг за подами можно делать либо нативно через `kubectl` либо интерактивно через [k9s](https://k9scli.io/topics/install/) (Я пользуюсь вторым, ибо он очень удобен):

```bash
cd $REPO_ROOT/infra/kubernetes/helm/values

helm repo add elastic https://helm.elastic.co

helm repo update

kubectl create namespace logging

helm install elasticsearch elastic/elasticsearch \
--namespace logging \
-f elasticsearch-values.yaml

helm install logstash elastic/logstash \
--namespace logging \
-f logstash-values.yaml

helm install kibana elastic/kibana \
--namespace logging \
-f kibana-values.yaml

helm install filebeat elastic/filebeat \
--namespace logging \
-f filebeat-values.yaml

echo "kibana-example.local $(minikube ip)" | sudo tee -a /etc/hosts
```

## Запуск клиента

Все задеплоено. Тестируем:

```bash
sudo docker run -it --rm --network=host isnov/gemini-gateway:client
```

И там пишите любые промпты.

## Мониторинг и логирование

Заходим на хосты которые были написаны в коде и делаем что душе угодно

## Дальнейшая работа

1. Тудушки
2. Хочется добавить возможность думающих ответов
3. Добавить контекст в ответы. Сейча каждое сообщение - это сообщение без контекста.
4. HTTPS соединение для `Ingress`?
