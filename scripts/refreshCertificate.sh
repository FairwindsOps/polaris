secretLabel="$(kubectl get secret $SECRET_NAME -o jsonpath='{.metadata.labels.completed}')"
key="$(kubectl get secret $SECRET_NAME -o jsonpath='{.data.key.pem}')"
cert="$(kubectl get secret $SECRET_NAME -o jsonpath='{.data.cert.pem}')"
caKey="$(kubectl get secret $SECRET_NAME -o jsonpath='{.data.ca-key.pem}')"
caCert="$(kubectl get secret $SECRET_NAME -o jsonpath='{.data.ca-cert.pem}')"
if (( $(($(date -d "$(echo $caKey | openssl x509 -noout -enddate | cut -c 10-)" +%s) - $(date -d "now" +%s))) >= $(( 86400 * 90 )) ))
then 
    echo $caCert > ca.crt
    echo $caKey > ca.key
    if (( $(($(date -d "$(echo $cert | openssl x509 -noout -enddate | cut -c 10-)" +%s) - $(date -d "now" +%s))) >= $(( 86400 * 30 )) ))
    then 
        echo $cert > server.crt
        echo $key > server.key
    fi 
fi 
# TODO expired certs
if [ "$secretLabel" = "completed" ]
then
    echo "Secret is already updated"
    exit 0
fi
country=US
state=MA
location=Boston
org=fairwinds
ou=insights
cn=$SERVICE_NAME
subj="/C=$country/ST=$state/L=$location/O=$org/OU=$ou/CN=$cn"
if [ ! -f ca.key ]
then
    # Generate self signed root CA cert
    openssl req -nodes -x509 -newkey rsa:2048 -days 3650 -keyout ca.key -out ca.crt -subj "$subj"
fi

if [ ! -f server.key ]
then
    # Generate server cert to be signed
    openssl req -nodes -newkey rsa:2048 -keyout server.key -out server.csr -subj "$subj"
    # Sign the server cert
    openssl x509 -req -in server.csr -days 180 -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
fi

# Save certificates
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: $SECRET_NAME
  labels:
    completed: completed
type: Opaque
data:
  key.pem: "$(cat server.key | base64 -w 0)"
  cert.pem: "$(cat server.crt | base64 -w 0)"
  ca-cert.pem: "$(cat ca.crt | base64 -w 0)"
  ca-key.pem: "$(cat ca.key | base64 -w 0)"
EOF
# Update Validating Webhook Configuration

kubectl get validatingwebhookconfiguration $WEBHOOK_CONFIG -o yaml | sed "s/caBundle:.*$/caBundle: $(cat ca.crt | base64 -w 0)/g" | kubectl apply -f -
