import { Box, Center, Flex, Spacer } from "@hope-ui/solid"
import LoginForm from "./LoginForm";

function ViewContainer() {
    return (
        <Box maxH="calc(100vh - 60px)" paddingTop="25px" paddingBottom="25px" overflow="auto">
            <Flex>
                <Spacer />
                    <Center>
                        <LoginForm />
                    </Center>
                <Spacer />
            </Flex>
        </Box>
    )
}

export default ViewContainer