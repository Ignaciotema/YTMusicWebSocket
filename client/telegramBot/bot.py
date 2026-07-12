import json
import logging
import os

from dotenv import load_dotenv
import websocket
from telegram import Update, ForceReply, InlineKeyboardMarkup, InlineKeyboardButton, ParseMode
from telegram.ext import Updater, CommandHandler, MessageHandler, Filters, CallbackContext, CallbackQueryHandler

logger = logging.getLogger(__name__)

# Store bot screaming status
screaming = False

# Conexion WS con el servidor YTMusicWebSocket (rol: controller)
WS_URL = "ws://localhost:8080/ws"
_ws_conn = None

# Texto del menu del reproductor
PLAYER_MENU = "<b>Reproductor</b>\n\nElegí una acción:"

# Comandos soportados por el dispatcher del servidor (ver dispatcher.go)
BACK_BUTTON = "⏮ Back"
PLAYPAUSE_BUTTON = "⏯ Play/Pause"
NEXT_BUTTON = "⏭ Next"

PLAYER_MENU_MARKUP = InlineKeyboardMarkup([[
    InlineKeyboardButton(BACK_BUTTON, callback_data="previous"),
    InlineKeyboardButton(PLAYPAUSE_BUTTON, callback_data="playPause"),
    InlineKeyboardButton(NEXT_BUTTON, callback_data="next"),
]])


def get_ws_connection():
    """
    Devuelve una conexion WS activa, registrandose como controller si hace falta.
    """

    global _ws_conn

    if _ws_conn is None or not _ws_conn.connected:
        _ws_conn = websocket.create_connection(WS_URL)
        _ws_conn.send(json.dumps({"type": "register", "data": "controller"}))

    return _ws_conn


def send_command(action: str) -> None:
    """
    Envia un comando (previous/playPause/next) al servidor WS.
    """

    conn = get_ws_connection()
    conn.send(json.dumps({"type": "command", "data": action}))


def echo(update: Update, context: CallbackContext) -> None:
    """
    This function would be added to the dispatcher as a handler for messages coming from the Bot API
    """

    # Print to console
    print(f'{update.message.from_user.first_name} wrote {update.message.text}')

    if screaming and update.message.text:
        context.bot.send_message(
            update.message.chat_id,
            update.message.text.upper(),
            # To preserve the markdown, we attach entities (bold, italic...)
            entities=update.message.entities
        )
    else:
        # This is equivalent to forwarding, without the sender's name
        update.message.copy(update.message.chat_id)


def scream(update: Update, context: CallbackContext) -> None:
    """
    This function handles the /scream command
    """

    global screaming
    screaming = True


def whisper(update: Update, context: CallbackContext) -> None:
    """
    This function handles /whisper command
    """

    global screaming
    screaming = False


def menu(update: Update, context: CallbackContext) -> None:
    """
    Muestra el keyboard con los controles del reproductor (Back / Play-Pause / Next)
    """

    context.bot.send_message(
        update.message.from_user.id,
        PLAYER_MENU,
        parse_mode=ParseMode.HTML,
        reply_markup=PLAYER_MENU_MARKUP
    )


def button_tap(update: Update, context: CallbackContext) -> None:
    """
    Procesa el boton tocado y despacha el comando correspondiente al servidor WS
    """

    query = update.callback_query
    action = query.data

    try:
        send_command(action)
    except (OSError, websocket.WebSocketException) as e:
        logger.error("No se pudo enviar el comando '%s': %s", action, e)

    # Cierra la animacion de carga del boton
    query.answer()


def main() -> None:
    load_dotenv()
    updater = Updater(os.getenv("BOT_API_KEY"))

    # Get the dispatcher to register handlers
    # Then, we register each handler and the conditions the update must meet to trigger it
    dispatcher = updater.dispatcher

    # Register commands
    dispatcher.add_handler(CommandHandler("scream", scream))
    dispatcher.add_handler(CommandHandler("whisper", whisper))
    dispatcher.add_handler(CommandHandler("menu", menu))

    # Register handler for inline buttons
    dispatcher.add_handler(CallbackQueryHandler(button_tap))

    # Echo any message that is not a command
    dispatcher.add_handler(MessageHandler(~Filters.command, echo))

    # Start the Bot
    updater.start_polling()

    # Run the bot until you press Ctrl-C
    updater.idle()


if __name__ == '__main__':
    main()