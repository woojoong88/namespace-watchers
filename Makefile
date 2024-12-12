
KUBESPRAY_VERSION ?= v2.26.0

deploy-kubernetes:
	rm -rf build/kubespray-configs/group_vars
	mkdir -p workspace
	cd workspace; git clone https://github.com/kubernetes-sigs/kubespray.git -b $(KUBESPRAY_VERSION) kubespray || true
	cp build/kubespray-configs/inventory.ini workspace/kubespray/inventory/sample/inventory.ini
	cd workspace; virtualenv venv
	cd workspace && source venv/bin/activate && cd kubespray && pip3 install -r requirements.txt && ansible-playbook -b -i inventory/sample/inventory.ini \
		-e "{'override_system_hostname' : False, 'disable_swap' : True}" \
		-e "{'kubeadm_enabled': True}" \
		-e "{'dns_min_replicas' : 1}" \
		cluster.yml

clean:
	cp build/kubespray-configs/inventory.ini workspace/kubespray/inventory/sample/inventory.ini
	cd workspace; virtualenv venv
	cd workspace && source venv/bin/activate && cd kubespray && pip3 install -r requirements.txt && ansible-playbook --extra-vars "reset_confirmation=yes" -b -i inventory/sample/inventory.ini reset.yml || true
	rm -rf workspace