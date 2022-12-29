# Allowing execute permissions on each file
chmod +x ./network-setup/network.sh
chmod +x ./network-setup/scripts/.
chmod +x ./network-setup/organizations/.

#Entering the network directory
cd ./network-setup/

#Starting the network
./network.sh up
./network.sh createChannel
echo Creato il network ed il canale

#Deploying smart contract

#Exiting the network directory
cd ..
