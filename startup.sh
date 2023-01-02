# Allowing execute permissions on each file
chmod +x ./network-setup/network.sh
chmod +x ./network-setup/scripts/.
chmod +x ./network-setup/organizations/.

#Entering the network directory
cd ./network-setup/

#Starting the network
./network.sh up
./network.sh createChannel
echo ##########################################
echo Creato il network ed il canale
echo ##########################################

#Deploying smart contract
./network.sh deployCC -ccn contract1 -ccp ../blockchain-application/seller-view/chaincode -ccl go
echo ##########################################
echo Contratto "contract1" per seller deployato
echo ##########################################

./network.sh deployCC -ccn contract2 -ccp ../blockchain-application/certificator-view/chaincode -ccl go
echo ##########################################
echo Contratto "contract2" per certificator deployato
echo ##########################################

#Exiting the network directory
cd ..
