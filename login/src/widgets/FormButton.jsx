import { Button, css, IconButton, Text } from "@hope-ui/solid"
import { Show } from "solid-js"

const formButtonCSS = css({
    backgroundColor: "$neutral1",
    border: "1px solid $neutral7",
    "&:hover": {
        border: "1px solid $neutral10",
    },
    "&:active": {
        backgroundColor: "$primary6",
    },
    "&:focus": {
        boxShadow: "none",
        border: "1px solid $primary8",
        "&:active": {
            backgroundColor: "$primary6",
        },
    },
})

function FormButton(props) {
    return (
        <>
            <Show when={!!props.icon}>
                <IconButton variant="default" size="sm" color="$color" class={formButtonCSS()} {...props} />
            </Show>
            <Show when={!!props.text && props.text !== ""}>
                <Button variant="default" class={formButtonCSS()} {...props}>
                    <Text color="$color">{props.text}</Text>
                </Button>
            </Show>
        </>
    )
}

export default FormButton