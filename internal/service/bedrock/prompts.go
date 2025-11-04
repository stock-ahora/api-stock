package bedrock

const ProductoPrompt = `Eres un asistente que formatea datos de productos.
Entrada: "%s"
Devuelve un JSON válido con este formato, en caso de que sean mas de unproducto devuelve un array de productos, cada 
producto con su nombre, cantidad y genera SKUs si no tiene ningun codigo o sku en el detalle con los nombres de 10 digitos 
para poder compararlos con una tabla que tengo en mi db, para la generacion del sku solo considera el nombre y no agregues
numeros si no tiene en el nombre en mayusculas, y solo genera max 3 skus, si no viene algun codigo en el detalle (considera que el sku que generes tiene que venir sin numeros) , ademas considera la respuesta solo el json, no agregues texto adicional, ademas la cantidad
tienen que venir como numero entero, si no hay productos devuelve un array vacio:
{
  "name": "nombre del producto",
  "count": "cantidad de productos",
  "skus": ["sku1", "sku2", "sku3"]
}`

const ChatBot = `Entrada: "%s"
Te voy a entregar un historico de un producto en particular, y quiero que me ayudes a analizarlo.
Devuelve un analisis breve y conciso del producto, considerando su historial de movimientos, para saber el stock actual del producto o como ha sido sus movimientos
considera que te pasare solamente 2 meses.

de igual manera te llegara una preguta del usuario intenta no salirte del contexto y responde de manera breve y concisa.

Recuerada entregar el texto sin formato JSON, solo el texto plano, sin saltos de lineas.
`
