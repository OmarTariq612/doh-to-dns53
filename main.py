from flask import Flask, request, make_response
import socket
import base64

app = Flask(__name__)

padding = "==="  # maximum padding required


@app.route("/dns-query", methods=["GET", "POST"])
def dns_query():
    match request.method:
        case "GET":
            encoded_request = request.args.get("dns")
            dns_request = base64.b64decode(encoded_request + padding, altchars=b"-_")
        case "POST":
            dns_request = request.data

    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
        sock.connect(("8.8.8.8", 53))
        sock.sendall(dns_request)
        sock.settimeout(5)
        try:
            dns_response = sock.recv(512)
        except socket.timeout:
            print("timeout exceeded")
            response = make_response("Gateway Timeout", 504)
            return response
    response = make_response(dns_response, 200)
    response.headers["Content-Type"] = "application/dns-message"
    return response


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=443, ssl_context=("site.crt", "site.key"))
