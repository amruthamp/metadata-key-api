# keyplay


Usage

1. clone repo to YOUR_SERVICE_NAME
2. run the following cmd 

    ```git grep -lz keyplay | xargs -0 sed -i '' -e 's/keyplay/YOUR_SERVICE_NAME/g'```

3. Find and address all TODO comments in the code