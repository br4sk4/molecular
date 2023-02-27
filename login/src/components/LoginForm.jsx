import {Box, Center, css, Flex, Spacer, Text, VStack} from "@hope-ui/solid";
import FormButton from "../widgets/FormButton";
import FormInput from "../widgets/FormInput";
import $ from "jquery";
import {createSignal} from "solid-js";

const boxCSS = css({
    width: "400px",
    backgroundColor: "$neutral6",
    border: "solid 1px $neutral9",
    borderRadius: "$lg",
})

document.addEventListener("keypress", function(event) {
    if (event.key === "Enter") {
        event.preventDefault();
        $("#signInButton").click();
    }
});

async function fetchToken (url, username, password) {
    let response = await fetch(url, {
        method: "GET",
        headers: {
            "Authorization": "Basic " + btoa(username + ":" + password)
        }
    });
    return await response.json();
}

function LoginForm() {
    const [username, setUsername] = createSignal("");
    const handleUsernameInput = event => setUsername(event.target.value);

    const [password, setPassword] = createSignal("");
    const handlePasswordInput = event => setPassword(event.target.value);

    const queryString = window.location.search;
    const urlParams = new URLSearchParams(queryString);
    const redirectUri = urlParams.get('redirect_uri');
    const state = urlParams.get('state');

    const onSignInClicked = () => {
        fetchToken(
            "/api/auth/",
            username(),
            password()
        ).then((response) => {
            window.location.replace(redirectUri + "?code=" + response.code + "&state=" + state);
        }).catch((err) => console.log(err))
    }

    return (
        <Box class={boxCSS()}>
            <Box bgColor="$neutral8">
                <Center>
                    <Text css={{height: "50px", lineHeight: "50px"}}>Sign In</Text>
                </Center>
            </Box>
            <Box padding="20px">
                <VStack spacing="10px">
                    <FormInput placeholder="Username" value={username()} onInput={handleUsernameInput} />
                    <FormInput placeholder="Password" value={password()} onInput={handlePasswordInput} type="password" />
                    <Flex w="100%">
                        <Spacer/>
                        <Box>
                            <FormButton onClick={onSignInClicked} id="signInButton" text="Sign in" />
                        </Box>
                    </Flex>
                </VStack>
            </Box>
        </Box>
    )
}

export default LoginForm