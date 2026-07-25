# Contexto

a ideia desse documento é explicar o objetivo desse sistema, contextualizar oq eu quero
antes de ferramenta antes de tecnologias, estamos falando de soluçao, entrega de valor pro usuário

# objetivo

eu quero que eu (ou alunos de ciência da computação e cursos relacionado a desenvolvimento) possam ir para a faculdade, apenas com o celular, sem levar o notebook, e poderem desenvolver em um ambiente proprio, para nao precisar ficar logando contas pessoais nos computadores da faculdade, como github, whatsapp email, etc. e eles possam estudar, e desenvolver softwares, nao em niveis gigantescos, o usuario nao vai fazer teste de performance e de carga nas aplicacoes dentro do celular, não nesse nivel, mas ele vai poder ir para a aula de C, javascriot, html, react, java, golang, python. E criar programas e rodar para fins de estudo, compilar e ver, ele fazer um `npm run dev` no celular, e colocar o ip de celular no browser, e ver a aplicacao dele rodando, enquanto ele edita o codigo no `lazyvim` ou no `vscode` que tambem esta rodando no celular em outra porta. e ter 100% de controle das informacoes, nao precisar se preocupar em deslogar, ou contectar toda vida q troca de sala.

# publico

## estudantes
tanto eu, q nao sou mais tao leigo, mas para pessoas q esta iniciando o meu  "sonho", é ter um app (apk de internet, oficial da play store ou um TUI bonito dentro do termux) q ele vai apertar os botoes "configurar" > "iniciar" > "abrir VS Code na porta X" > usar > "parar" ao fim da aula, vai continuar no controle dele os dados, e voltar para casa e continuar,

## profissionais
Apesar te ter esse modo facil uso, eu quero poder q a pessoa ainda vai ter acesso ao host do ubuntu pelo celular.



# maior desafio

hoje o maior problema é a parte visual, e quando eu falo isso é a visualizacao e uso sem ser por terminal,
o sonho impossivel, seria um GUI completo, mas nao tem como pq trava muito, entao a experiencia desktop com GUI nao dá
Dentro do gui, eu diria que tem coisas q nao daria pra descartar, q seria o browser, o mundo ideal dentro da realidade  seria poder fazer um `firefox start port=8080` mas nao é assim pq basicamente no mundo moderno, oq nao da pra fazer por TUI da comunidade, daria pra ver dentro do browser.

sabendo q isso é extremamente dificil, a solução seria todar individualmente os apps em suas portas, um pro vscode, um pra outra coisa, etc.

entao para experiencias desktop. teria q ser o "Desktop-TUI" ou o "VTM". se tivesse algo assim q tivesse acesso ao ambiente linux, e renderizasse em um site, "tipo um Desktop-TUI feito em next/react" seria talvez até mais perfeito q o GUI puro por ser mais leve

Mas dito tudo isso, o creio q o MVP 1 seria ser 100% TUI, tudo exposto por uma unica porta SSH, aonde o usuario teria o acesso completo por ali, mais nada. Sem porta extra pro code-server. somente o porta do ssh, e claro, os proprios apps do usuario,tipo o um nom run dev ou go run

