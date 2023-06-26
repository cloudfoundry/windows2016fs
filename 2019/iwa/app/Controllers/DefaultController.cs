using System.Data.SqlClient;
using System.Runtime.Remoting.Messaging;
using System;
using System.Threading;
using System.Web.Http;

namespace WindowsAuth.Controllers
{
    public class DefaultController : ApiController
    {
        [Route("~/")]
        [AllowAnonymous]
        [HttpGet]
        public IHttpActionResult Root()
        {
            return Ok("Howdy! I am a Windows Authentication test app!");
        }

        [Route("~/auth")]
        [HttpGet]
        [Authorize]
        public IHttpActionResult Auth()
        {
            var identity = Thread.CurrentPrincipal.Identity;
            return Ok($"Logged in as {identity.Name} via method {identity.AuthenticationType}.");
        }
        [Route("~/sql")]
        [HttpGet]
        //[Authorize]
        public IHttpActionResult Sql()
        {
            
            var identity = Thread.CurrentPrincipal.Identity;
            var cs = System.Configuration.ConfigurationManager.ConnectionStrings["mydb"].ConnectionString;
            Console.WriteLine(cs);
            string queryString = "SELECT * FROM MyTest;";
            using (SqlConnection connection = new SqlConnection(cs))
            {
                SqlCommand command = new SqlCommand(queryString, connection);
                connection.Open();
                SqlDataReader reader = command.ExecuteReader();
                try
                {
                    while (reader.Read())
                    {
                        Console.WriteLine(String.Format("{0}, {1}", reader["Name"], reader["Age"]));
                        return Ok(String.Format("{0}, {1}", reader["Name"], reader["Age"]));
                    }
                }
                finally
                {
                        // Always call Close when done reading.
    reader.Close();
                    }
            }

            return Ok("Failed to read");
        }

        [Route("~/sql-login")]
        [HttpGet]
        //[Authorize]
        public IHttpActionResult SqlLogin()
        {

            var identity = Thread.CurrentPrincipal.Identity;
            var cs = System.Configuration.ConfigurationManager.ConnectionStrings["mydb-sql-login"].ConnectionString;
            Console.WriteLine(cs);
            string queryString = "SELECT * FROM MyTest;";
            using (SqlConnection connection = new SqlConnection(cs))
            {
                SqlCommand command = new SqlCommand(queryString, connection);
                connection.Open();
                SqlDataReader reader = command.ExecuteReader();
                try
                {
                    while (reader.Read())
                    {
                        Console.WriteLine(String.Format("{0}, {1}", reader["Name"], reader["Age"]));
                        return Ok(String.Format("{0}, {1}", reader["Name"], reader["Age"]));
                    }
                }
                finally
                {
                    // Always call Close when done reading.
                    reader.Close();
                }
            }

            return Ok("Failed to read");
        }

    }
}
