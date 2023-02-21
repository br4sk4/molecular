import {Box, Flex, Heading, HStack, Image, Text} from "@hope-ui/solid"
import moleculeImage from '../assets/molecule.png'

function AppBar() {
    return (
        <Box height="60px" width="100vw">
            <Flex bg="$primary6" css={{ padding: "10px" }}>
                <Box p="$1" color="white" paddingTop="6px">
                    <Heading size="xl" fontWeight="$bold">
                        <HStack spacing="15px">
                            <Image boxSize="30px" src={moleculeImage} alt="molecule" objectFit="cover" />
                            <Text css={{height: "30px", lineHeight: "30px"}}>Molecular - Auth Proxy</Text>
                        </HStack>
                    </Heading>
                </Box>
            </Flex>
        </Box>
    )
}

export default AppBar