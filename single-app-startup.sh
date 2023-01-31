# Allowing execute permissions on each file
chmod +x ./network-setup/network.sh
chmod +x ./network-setup/scripts/.
chmod +x ./network-setup/organizations/.

#Entering the network directory
cd ./network-setup/

#Starting the network
./network.sh up
./network.sh createChannel


#Deploying smart contract
./network.sh deployCC -ccn contract1 -ccp ../blockchain-application/general-view/chaincode/ -ccl go


#Exiting the network directory
cd ..
