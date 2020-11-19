import discord, json, os

print("Enter Discord Token")
Token = input()

class Client(discord.Client):

    def __init__(self, **options):
        (super().__init__)(**options)

    def task(self, guildid: int):
        guild = self.get_guild(guildid)
        memberslist = []
        for member in guild.members:
            memberslist.append(str(member.id))

        jsonified = json.dumps(memberslist, separators=(',', ':'))

        with open('members.json', 'w') as (f):
            f.write(jsonified)


    async def on_connect(self):
        print('Scrape Ready => ' + client.user.name)

    async def on_message(self, message):
        if message.author == self.user:
            self.task(message.guild.id)


if __name__ == '__main__':
    client = Client()
    client.run(Token, bot=False)