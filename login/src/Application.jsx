import {Box} from "@hope-ui/solid";
import AppBar from "./components/AppBar";
import ViewContainer from "./components/ViewContainer";

function Application() {
    return (
        <Box width="100vw" height="100vh">
            <AppBar />
            <ViewContainer />
        </Box>
    )
}

export default Application